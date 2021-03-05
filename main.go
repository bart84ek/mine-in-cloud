package main

import (
	"log"

	"b4rt.io/aws-mine-manager/cloud"
	"b4rt.io/aws-mine-manager/mine"
)

func main() {
	awsCloud, err := cloud.NewAWS()
	if err != nil {
		log.Fatal("cloud error", err.Error())
	}

	mineMgr, err := mine.NewManager(awsCloud)
	if err != nil {
		log.Fatal("mine manager error", err.Error())
	}

	instances, err := mineMgr.GetMines()
	for _, mine := range instances {
		log.Println("MineId:", mine.Id, "; IP:", mine.PublicIpAddress, ";State:", mine.State)
	}

	// instances[0].Connect()
	// skip if got running mine ?
	// CreateMine(mineMgr)
	log.Println("Done")
}

func CreateMine(mineMgr mine.MineManager) {
	imageId := "ami-0767046d1677be5a0"
	keyName := "minecraftKeysNew"
	securityGroup := "launch-wizard-3"
	mineMgr.CreateMine(imageId, keyName, securityGroup)
}
