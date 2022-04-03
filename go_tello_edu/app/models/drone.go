package models

import (
	"io"
	"log"
	"os/exec"
	"strconv"
	"time"

	"github.com/hybridgroup/mjpeg"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"gocv.io/x/gocv"
	"golang.org/x/sync/semaphore"
)

const (
	DefaultSpeed      = 10
	WaitDroneStartSec = 5
	frameX            = 960 / 3
	frameY            = 720 / 3
	frameCenterX      = frameX / 2
	frameCenterY      = frameY / 2
	frameArea         = frameX * frameY
	frameSize         = frameArea * 3
)

type DroneManager struct {
	*tello.Driver
	Speed        int
	patrolSem    *semaphore.Weighted
	patrolQuit   chan bool
	isPatrolling bool
	// pipe0でドローンのvideoを書き込む
	ffmpegIn io.WriteCloser
	// pipe1でドローンのvideoを読み込む
	ffmpegOut io.ReadCloser
	Stream    *mjpeg.Stream
}

func NewDroneManager() *DroneManager {
	drone := tello.NewDriverWithIP("192.168.10.1", "8888")

	ffmpeg := exec.Command("ffmpeg", "-hwaccel", "auto", "-hwaccel_device", "opencl", "-i", "pipe:0", "-pix_fmt", "bgr24",
		"-s", strconv.Itoa(frameX)+"x"+strconv.Itoa(frameY), "-f", "rawvideo", "pipe:1")
	ffmpegIn, _ := ffmpeg.StdinPipe()
	ffmpegOut, _ := ffmpeg.StdoutPipe()

	droneManager := &DroneManager{
		Driver:       drone,
		Speed:        DefaultSpeed,
		patrolSem:    semaphore.NewWeighted(1),
		patrolQuit:   make(chan bool),
		isPatrolling: false,
		ffmpegIn:     ffmpegIn,
		ffmpegOut:    ffmpegOut,
		Stream:       mjpeg.NewStream(),
	}
	work := func() {
		if err := ffmpeg.Start(); err != nil {
			log.Println(err)
			return
		}

		drone.On(tello.ConnectedEvent, func(data interface{}) {
			log.Println("Connected")
			drone.StartVideo()
			drone.SetVideoEncoderRate(tello.VideoBitRateAuto)
			drone.SetExposure(0)

			gobot.Every(100*time.Millisecond, func() {
				drone.StartVideo()
			})

			droneManager.StreamVideo()
		})

		drone.On(tello.VideoFrameEvent, func(data interface{}) {
			pkt := data.([]byte)
			if _, err := ffmpegIn.Write(pkt); err != nil {
				log.Println(err)
			}
		})
	}
	robot := gobot.NewRobot("tello", []gobot.Connection{}, []gobot.Device{drone}, work)
	// goroutineを使わないと以降のコードが実行されない
	// ->非同期に実行
	go robot.Start()
	// goroutineを使うとドローンとコネクションできているか確認できない
	// コネクションしない状態でtakeoffなどを呼ぶと、invalid memory errorが出る可能性あり
	time.Sleep(WaitDroneStartSec * time.Second)
	return droneManager
}

// 巡回と停止を兼ねている
func (d *DroneManager) Patrol() {
	go func() {
		isAquire := d.patrolSem.TryAcquire(1)
		if !isAquire {
			d.patrolQuit <- true
			d.isPatrolling = false
			log.Println("パトロール終了")
			return
		}
		defer d.patrolSem.Release(1)

		log.Println("パトロール開始")
		d.isPatrolling = true
		status := 0
		t := time.NewTicker(3 * time.Second)

		for {
			select {
			case <-t.C:
				d.Hover()
				switch status {
				case 1:
					log.Println("前")
					d.Forward(d.Speed)
				case 2:
					log.Println("右")
					d.Right(d.Speed)
				case 3:
					log.Println("後")
					d.Backward(d.Speed)
				case 4:
					log.Println("左")
					d.Left(d.Speed)
				case 5:
					status = 0
				}
				status++
			case <-d.patrolQuit:
				t.Stop()
				d.Hover()
				d.isPatrolling = false
				return
			}
		}
	}()
}

func (d *DroneManager) StartPatrol() {
	if !d.isPatrolling {
		d.Patrol()
	}
}

func (d *DroneManager) StopPatrol() {
	if d.isPatrolling {
		d.Patrol()
	}
}

func (d *DroneManager) StreamVideo() {
	go func(d *DroneManager) {
		for {
			buf := make([]byte, frameSize)
			if _, err := io.ReadFull(d.ffmpegOut, buf); err != nil {
				log.Println(err)
			}
			img, _ := gocv.NewMatFromBytes(frameY, frameX, gocv.MatTypeCV8UC3, buf)

			if img.Empty() {
				continue
			}

			// IMEncodeの返り値がバイト配列から*NativeByteBufferになったため、コードを変更
			// https://github.com/hybridgroup/gocv/commit/5dbdee404ae6dff1e291080c80973ffd1abdd056
			jpegBuf, _ := gocv.IMEncode(".jpg", img)
			jpegBytes := jpegBuf.GetBytes()

			d.Stream.UpdateJPEG(jpegBytes)
		}
	}(d)
}
