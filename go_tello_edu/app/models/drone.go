package models

import (
	"log"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"golang.org/x/sync/semaphore"
)

const (
	DefaultSpeed      = 10
	WaitDroneStartSec = 5
)

type DroneManager struct {
	*tello.Driver
	Speed        int
	patrolSem    *semaphore.Weighted
	patrolQuit   chan bool
	isPatrolling bool
}

func NewDroneManager() *DroneManager {
	drone := tello.NewDriver("8889")
	droneManager := &DroneManager{
		Driver:       drone,
		Speed:        DefaultSpeed,
		patrolSem:    semaphore.NewWeighted(1),
		patrolQuit:   make(chan bool),
		isPatrolling: false,
	}
	work := func() {
		// TODO
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
		log.Println("パトロールstart")
		isAquire := d.patrolSem.TryAcquire(1)
		if !isAquire {
			d.patrolQuit <- true
			d.isPatrolling = false
			return
		}
		defer d.patrolSem.Release(1)
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
