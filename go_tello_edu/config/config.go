package config

import (
	"log"
	"os"

	"gopkg.in/ini.v1"
)

// type ConfList struct {
// 	LogFile string
// 	Address string
// 	Port    string
// }
//
// var Config ConfList
//
// func init() {
// 	cfg, err := ini.Load("config.ini")
// 	if err != nil {
// 		log.Printf("Failde to read file: %w", err)
// 		os.Exit(1)
// 	}
//
// 	Config = ConfList{
// 		LogFile: cfg.Section("go tello").Key("log_file").String(),
// 		Address: cfg.Section("web").Key("address").String(),
// 		Port:    cfg.Section("web").Key("port").String(),
// 	}
// }
//

type ConfigList struct {
	LogFile string
	Address string
	Port    int
}

var Config ConfigList

func init() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Printf("Failed to read: %v", err)
		os.Exit(1)
	}
	Config = ConfigList{
		LogFile: cfg.Section("go_tello_edu").Key("log_file").String(),
		Address: cfg.Section("web").Key("address").String(),
		Port:    cfg.Section("web").Key("port").MustInt(),
	}
}
