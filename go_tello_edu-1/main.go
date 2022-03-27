package main

import (
	"time"
	"web_app/go_tello_edu-1/app/models"
	"web_app/go_tello_edu-1/app/utils"
	"web_app/go_tello_edu-1/config"
)

func main() {
	utils.LoggingSettings(config.Config.LogFile)
	droneManager := models.NewDroneManager()
	droneManager.TakeOff()
	time.Sleep(10 * time.Second)
	droneManager.Land()
}
