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
	Speed       int
	patrolSem   *semaphore.Weighted
	patrolQuit  chan bool
	isPatroling bool
}

func NewDroneManager() *DroneManager {
	drone := tello.NewDriver("8889")
	droneManager := &DroneManager{
		Driver:      drone,
		Speed:       DefaultSpeed,
		patrolSem:   semaphore.NewWeighted(1),
		patrolQuit:  make(chan bool),
		isPatroling: false,
	}
	work := func() {
		// TODO
	}
	robot := gobot.NewRobot("tello", []gobot.Connection{}, []gobot.Device{}, work)
	go robot.Start()
	time.Sleep(WaitDroneStartSec * time.Second)

	return droneManager
}

func (d *DroneManager) Patrol() {
	isAcquire := d.patrolSem.TryAcquire(1)
	if !isAcquire {
		return
	}

	defer d.patrolSem.Release(1)
	d.isPatroling = true
	status := 0
	t := time.NewTicker(3 * time.Second)

	for {
		select {
		case <-t.C:
			d.Hover()
			switch status {
			case 1:
				d.Forward(d.Speed)
			case 2:
				d.Right(d.Speed)
			case 3:
				d.Backward(d.Speed)
			case 4:
				d.Left(d.Speed)
			}
		}
	}
}
