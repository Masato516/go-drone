package models

import (
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
		Driver: drone,
		Speed:  DefaultSpeed,
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
