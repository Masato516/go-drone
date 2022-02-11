package main

import (
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
)

func main() {
	drone := tello.NewDriverWithIP("192.168.10.1", "8889")

	work := func() {
		drone.TakeOff()

		gobot.After(10*time.Second, func() {
			drone.BackFlip()
		})

		gobot.After(15*time.Second, func() {
			drone.Hover()
		})

		gobot.After(20*time.Second, func() {
			drone.FrontFlip()
		})

		gobot.After(25*time.Second, func() {
			drone.Hover()
		})

		gobot.After(30*time.Second, func() {
			drone.LeftFlip()
		})

		gobot.After(35*time.Second, func() {
			drone.Hover()
		})

		gobot.After(40*time.Second, func() {
			drone.RightFlip()
		})

		gobot.After(45*time.Second, func() {
			drone.Hover()
		})

		gobot.After(50*time.Second, func() {
			drone.Land()
		})
	}

	robot := gobot.NewRobot("tello",
		[]gobot.Connection{},
		[]gobot.Device{drone},
		work,
	)

	robot.Start()
}

//import "net"
//
//func main() {
//	conn, _ := net.Dial("udp", "192.168.10.1:8889")
//	conn.Write([]byte("command"))
//	conn.Write([]byte("takeoff"))
//}
//
