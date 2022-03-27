package main

import (
	"log"
	"udemy_drone/go_tello_edu/app/controllers"
	"udemy_drone/go_tello_edu/config"
	"udemy_drone/go_tello_edu/utils"
)

func main() {
	// utils.LoggingSettings(config.Config.LogFile)
	// droneManager := models.NewDroneManager()
	// droneManager.Land()
	// droneManager.TakeOff()
	// time.Sleep(10 * time.Second)
	// droneManager.Patrol()
	// log.Println("パトロール１回目終了")
	// time.Sleep(30 * time.Second)
	// droneManager.Patrol()
	// time.Sleep(10 * time.Second)
	// droneManager.Land()

	utils.LoggingSettings(config.Config.LogFile)
	log.Println(controllers.StartWebServer())
}
