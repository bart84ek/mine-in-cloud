package main

import (
	"fmt"
	"log"
	"time"

	"b4rt.io/aws-mine-manager/cloud"
	"b4rt.io/aws-mine-manager/config"
	"b4rt.io/aws-mine-manager/mine"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatal("config error", err.Error())
	}

	awsCloud, err := cloud.AWS()
	if err != nil {
		log.Fatal("cloud error", err.Error())
	}

	mineMgr, err := mine.NewManager(awsCloud, cfg.AWSSSHUser, cfg.AWSSSSPort, cfg.AWSKeyPath)
	if err != nil {
		log.Fatal("mine manager error", err.Error())
	}

	mines, err := mineMgr.GetMines()
	for _, mine := range mines {
		log.Println("MineId:", mine.Id, "; IP:", mine.PublicIpAddress, ";State:", mine.State)
	}

	log.Println(fmt.Sprintf("Found %d mine(s)", len(mines)))

	if len(mines) == 0 {
		log.Println("Creating new mine")

		mine, err := mineMgr.CreateMine(cfg.AWSImageId, cfg.AWSKeyName, cfg.AWSSecurityGroup)
		if err != nil {
			log.Fatal("error creating mine", err)
		}

		log.Printf("New mine. Id:%s IP:%s (vm:%s, ssh:notrunning). (wait 5 sec)", mine.Id, mine.PublicIpAddress, mine.State)

		for !mine.IsReady() {
			log.Printf("Mine not ready. Id:%s IP:%s (vm:%s, ssh:notrunning). (wait 5 sec)", mine.Id, mine.PublicIpAddress, mine.State)
			time.Sleep(5 * time.Second)
			mine, err = mineMgr.GetMine(mine.Id)
			if err != nil {
				log.Fatal("error finding mine", err)
			}
		}
		log.Println("Setting up mine")

		err = mine.Setup()
		if err != nil {
			log.Fatal("error while setup mine ", err)
		}
	} else {
		log.Println("Found mine", mines[0].Id)
		err := mines[0].Backup()
		if err != nil {
			log.Fatal("error while setup mine ", err)
		}
	}

}
