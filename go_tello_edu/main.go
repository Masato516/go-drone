package main

import (
	"log"
	"udemy_drone/go_tello_edu/app/utils"
	"udemy_drone/go_tello_edu/config"
)

func main() {
	utils.LoggingSettings(config.Config.LogFile)
	log.Println("test")
	// droneManager := models.NewDroneManager()
	// droneManager.TakeOff()
	// time.Sleep(10 * time.Second)
	// droneManager.Land()
}
