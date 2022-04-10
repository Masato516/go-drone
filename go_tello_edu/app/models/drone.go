package models

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"math"
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
	faceDetectXMLFile = "app/models/haarcascade_frontalface_default.xml"
	snapshotsFolder   = "static/img/snapshots/"
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
	ffmpegOut            io.ReadCloser
	Stream               *mjpeg.Stream
	faceDetectTrackingOn bool
	isSnapShot           bool
}

func NewDroneManager() *DroneManager {
	drone := tello.NewDriverWithIP("192.168.10.1", "8888")

	ffmpeg := exec.Command("ffmpeg", "-hwaccel", "auto", "-hwaccel_device", "opencl", "-i", "pipe:0", "-pix_fmt", "bgr24",
		"-s", strconv.Itoa(frameX)+"x"+strconv.Itoa(frameY), "-f", "rawvideo", "pipe:1")
	ffmpegIn, _ := ffmpeg.StdinPipe()
	ffmpegOut, _ := ffmpeg.StdoutPipe()

	droneManager := &DroneManager{
		Driver:               drone,
		Speed:                DefaultSpeed,
		patrolSem:            semaphore.NewWeighted(1),
		patrolQuit:           make(chan bool),
		isPatrolling:         false,
		ffmpegIn:             ffmpegIn,
		ffmpegOut:            ffmpegOut,
		Stream:               mjpeg.NewStream(),
		faceDetectTrackingOn: false,
		isSnapShot:           false,
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

		drone.Once(tello.FlightDataEvent, func(data interface{}) {
			pkt := data.(*tello.FlightData)
			log.Println("Battery:", pkt.BatteryPercentage, "%")
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

// 画像への装飾を加えるメソッド
func (d *DroneManager) StreamVideo() {
	go func(d *DroneManager) {
		classifier := gocv.NewCascadeClassifier()
		defer classifier.Close()

		if !classifier.Load(faceDetectXMLFile) {
			log.Printf("Error reading cascade file: %s", faceDetectXMLFile)
			return
		}

		// color for the rect when faces detected
		blue := color.RGBA{0, 0, 255, 0}

		for {
			buf := make([]byte, frameSize)
			if _, err := io.ReadFull(d.ffmpegOut, buf); err != nil {
				log.Println(err)
			}
			img, _ := gocv.NewMatFromBytes(frameY, frameX, gocv.MatTypeCV8UC3, buf)

			if img.Empty() {
				continue
			}

			if d.faceDetectTrackingOn {
				d.StopPatrol()
				// detect faces
				rects := classifier.DetectMultiScale(img)
				fmt.Printf("found %d faces\n", len(rects))
				// 顔が検出されない場合は、一時停止
				if len(rects) == 0 {
					fmt.Println("顔が見つかりません")
					d.Hover()
				}

				// draw a rectangle around each face on the original image
				for _, r := range rects {
					gocv.Rectangle(&img, r, blue, 3)
					// Pt is shorthand for Point{X, Y}
					pt := image.Pt(r.Max.X, r.Min.Y-5)
					gocv.PutText(&img, "Human", pt, gocv.FontHersheyPlain, 1.2, blue, 2)
					// 顔を追跡する
					d.chaseFace(r)
					break // 認識する顔を１つに留める
				}
			}

			// IMEncodeの返り値がバイト配列から*NativeByteBufferになったため、コードを変更
			// https://github.com/hybridgroup/gocv/commit/5dbdee404ae6dff1e291080c80973ffd1abdd056
			jpegBuf, _ := gocv.IMEncode(".jpg", img)
			jpegBytes := jpegBuf.GetBytes()

			if d.isSnapShot {
				log.Println("スナップショットが保存されました")
				backupFileName := snapshotsFolder + time.Now().Format(time.RFC3339) + ".jpg"
				err := ioutil.WriteFile(backupFileName, jpegBytes, 0644)
				if err != nil {
					log.Printf("cannot save snapshot: %s", err.Error())
				}
				snapshotFileName := snapshotsFolder + "snapshot.jpg"
				ioutil.WriteFile(snapshotFileName, jpegBytes, 0644)
				d.isSnapShot = false
			}

			d.Stream.UpdateJPEG(jpegBytes)
		}
	}(d)
}

func (d *DroneManager) EnableFaceDetectTracking() {
	d.faceDetectTrackingOn = true
}

func (d *DroneManager) DisableFaceDetectTracking() {
	d.faceDetectTrackingOn = false
	d.Hover()
}

func (d *DroneManager) chaseFace(r image.Rectangle) {
	move := false
	// 前後左右での追跡
	faceCenterX := (r.Max.X + r.Min.Y) / 2
	faceCenterY := (r.Max.Y + r.Min.Y) / 2
	diffX := frameCenterX - faceCenterX
	diffY := frameCenterY - faceCenterY

	if diffX > 20 {
		d.Left(10)
		fmt.Println("左に移動")
		move = true
	}
	if diffX < -20 {
		d.Right(10)
		fmt.Println("右に移動")
		move = true
	}
	if diffY > 30 {
		d.Up(10)
		fmt.Println("上に移動")
		move = true
	}
	if diffY < -30 {
		d.Down(10)
		fmt.Print("下に移動")
		move = true
	}

	// 前後での追跡
	faceWidth := r.Max.X - r.Min.X
	faceHeight := r.Max.Y - r.Min.Y
	faceArea := faceWidth * faceHeight
	percentF := math.Round(float64(faceArea) / float64(frameArea) * 100)
	fmt.Println(percentF)

	if percentF > 15 {
		d.Backward(10)
		fmt.Println("後ろに移動")
		move = true
	}
	if percentF < 5 {
		d.Forward(10)
		fmt.Println("前に移動")
		move = true
	}

	if !move {
		d.Hover()
	}
}

func (d *DroneManager) TakeSnapshot() {
	d.isSnapShot = true
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// 2秒経っても処理が終わらなければ中断する
	for {
		if !d.isSnapShot || ctx.Err() != nil {
			break
		}
	}
	d.isSnapShot = false
}
