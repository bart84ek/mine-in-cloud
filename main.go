package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"b4rt.io/aws-mine-manager/cloud"
	"b4rt.io/aws-mine-manager/config"
	"b4rt.io/aws-mine-manager/mine"
)

var commands = map[string]func(config.Config){
	"list":    listCmd,
	"list-ip": listAddressesCmd,
	"start":   createMineCmd,
	"stop":    stopMineCmd,
	"players": listPlayers,
}

func main() {
	if len(os.Args) < 2 {
		showUsage()
	}

	cmd := os.Args[1]
	if command, ok := commands[cmd]; ok {
		cfg, err := config.LoadFromEnv()
		if err != nil {
			log.Fatal("config error", err.Error())
		}
		command(cfg)
	} else {
		if cmd != "help" {
			log.Printf("Unknown command \"%s\"", cmd)
		}
		showUsage()
		os.Exit(1)
	}
}

func showUsage() {
	keys := make([]string, 0, len(commands))
	for k := range commands {
		keys = append(keys, k)
	}
	log.Printf("Usage: %s", strings.Join(keys, " | "))
}

func listCmd(cfg config.Config) {
	mineMgr, err := getMineManager(cfg)
	if err != nil {
		log.Fatal("mine manager error", err.Error())
	}

	mines, err := mineMgr.GetMines()
	if err != nil {
		log.Fatal("error fetching mines list ", err.Error())
	}

	log.Println(fmt.Sprintf("Found %d mine(s)", len(mines)))
	for _, mine := range mines {
		log.Println("MineId:", mine.Id, "; IP:", mine.PublicIpAddress, ";State:", mine.State)
	}
}

func listPlayers(cfg config.Config) {
	mineMgr, err := getMineManager(cfg)
	if err != nil {
		log.Fatal("mine manager error", err.Error())
	}

	mines, err := mineMgr.GetMines()
	if err != nil {
		log.Fatal("error fetching mines list ", err.Error())
	}

	if len(mines) < 1 {
		log.Fatal("No mine found")
	}

	mine := mines[0]

	log.Println("Players minecraft server")
	err = mine.Players()
	if err != nil {
		log.Fatal("Warning. Can't fetch list of players. Minecraft Server running?")
	}
}

func listAddressesCmd(cfg config.Config) {
	mineMgr, err := getMineManager(cfg)
	if err != nil {
		log.Fatal("mine manager error", err.Error())
	}
	mineMgr.GetAddresses()
}

func createMineCmd(cfg config.Config) {
	mineMgr, err := getMineManager(cfg)
	if err != nil {
		log.Fatal("mine manager error", err.Error())
	}

	log.Println("Creating new mine")
	mine, err := mineMgr.CreateMine(cfg.AWS.ImageId, cfg.AWS.Key.Name, cfg.AWS.SecurityGroup)
	if err != nil {
		log.Fatal("error creating mine", err)
	}

	log.Printf("New mine. Id:%s IP:%s (vm:%s, ssh:notrunning).", mine.Id, mine.PublicIpAddress, mine.State)

	for !mine.IsReady() {
		log.Printf("Mine not ready. Id:%s IP:%s (vm:%s, ssh:notrunning). (wait 5 sec)", mine.Id, mine.PublicIpAddress, mine.State)
		time.Sleep(5 * time.Second)
		mine, err = mineMgr.GetMine(mine.Id)
		if err != nil {
			log.Fatal("error finding mine", err)
		}
	}
	log.Println("Setting up mine")

	setupFiles := strings.Split(cfg.Mine.SetupFiles, " ")
	setupCmds := strings.Split(cfg.Mine.SetupCmds, " ")
	err = mine.Setup(
		cfg.Mine.SetupDirPath, setupFiles, setupCmds,
		cfg.Mine.BackupDirPath, cfg.Mine.LatestDataFileName, cfg.Mine.DataDirPath)
	if err != nil {
		log.Fatal("error setup mine ", err)
	}

	log.Println("Assigning ElasticIP")
	err = mineMgr.AssignElasticIP(mine.Id)
	if err != nil {
		log.Fatal("Error during assigning ElasticIP", err)
	}
}

func stopMineCmd(cfg config.Config) {
	mineMgr, err := getMineManager(cfg)
	if err != nil {
		log.Fatal("mine manager error", err.Error())
	}

	mines, err := mineMgr.GetMines()
	if err != nil {
		log.Fatal("error fetching mines list ", err.Error())
	}

	if len(mines) < 1 {
		log.Fatal("No mine found")
	}

	mine := mines[0]

	log.Println("Players minecraft server")
	err = mine.Players()
	if err != nil {
		log.Println("Warning. Can't fetch list of players. Minecraft Server running?")
	}

	log.Println("Stopping minecraft server")
	err = mine.Save()
	if err != nil {
		log.Println("Warning. Can't save current state Minecraft Server running?", err.Error())
	}

	log.Println("Backup mine")
	err = mine.Backup(cfg.Mine.DataDirPath, cfg.Mine.BackupDirPath, cfg.Mine.BackupFileFormat, cfg.Mine.LatestDataFileName)
	if err != nil {
		log.Fatal("Error during mine backup", err.Error())
	}

	log.Println("Terminate mine. Bye bye")
	err = mineMgr.Terminate(mine.Id)
	if err != nil {
		log.Fatal("Error during mine Termination", err)
	}

	log.Println("Releasing ElasticIP")
	err = mineMgr.ReleaseElasticIP()
	if err != nil {
		log.Fatal("Error during releasing ElasticIP", err)
	}
}

func getMineManager(cfg config.Config) (mine.MineManager, error) {
	awsCloud, err := cloud.AWS()
	if err != nil {
		log.Fatal("cloud error", err.Error())
	}

	awsCfg := cfg.AWS
	return mine.NewManager(awsCloud, awsCfg.SSH.User, awsCfg.SSH.Port, awsCfg.Key.Path, awsCfg.ElasticIP.AllocID, awsCfg.ElasticIP.ReleaseToNetInterfaceID)
}
