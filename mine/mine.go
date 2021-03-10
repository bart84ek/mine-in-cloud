package mine

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	ssh "github.com/helloyi/go-sshclient"
)

type Mine struct {
	Id              string
	PublicIpAddress string
	State           string
	SSHUsername     string
	SSHKeyPath      string
	SSHPort         int
}

var filesToCopy = []string{
	"setup.sh",
	"runMinecraftServer.sh",
	"/Users/bart/poligon/aws/aws-mine-manager/backups/latest.tar.gz",
}

var setupCmds = []string{
	"./setup.sh",
	"tar zxf latest.tar.gz",
	"./runMinecraftServer.sh",
}

func (m *Mine) IsReady() bool {
	if m.State != "running" {
		return false
	}
	if !sshIsUp(m) {
		return false
	}
	return true
}

func (m *Mine) Setup() error {
	log.Println("Mine setup")
	log.Println("Copying setup files.")
	err := scpFilesToMine(m, filesToCopy)
	if err != nil {
		return err
	}

	log.Println("Running setup commands.")
	client, err := sshClient(m)
	if err != nil {
		return err
	}
	defer client.Close()

	for _, setupCmd := range setupCmds {
		output, err := client.Cmd(setupCmd).SmartOutput()
		log.Println(string(output))
		if err != nil {
			return err
		}

	}
	return nil
}

func (m *Mine) Backup() error {
	client, err := sshClient(m)
	if err != nil {
		return err
	}
	defer client.Close()

	cmds := []string{
		"sudo docker exec minecraft-server rcon-cli list",
		"sudo docker exec minecraft-server rcon-cli save-all",
		// "sudo docker exec minecraft-server rcon-cli stop",
		"tar czf data.tar.gz data/",
	}

	for _, cmd := range cmds {
		out, err := client.Cmd(cmd).SmartOutput()
		log.Println(string(out))
		if err != nil {
			return err
		}
	}

	now := time.Now()
	timeStr := fmt.Sprintf("%02d-%02d-%dT%02d%02d", now.Day(), now.Month(), now.Year(), now.Hour(), now.Minute())
	backupFileName := fmt.Sprintf("data_%s.tar.gz", timeStr)
	backupFilePath := fmt.Sprintf("backups/%s", backupFileName)
	backupLatestPath := "backups/latest.tar.gz"

	err = scpFileFromMine(m, "~/data.tar.gz", backupFilePath)
	if err != nil {
		return err
	}

	os.Remove(backupLatestPath)

	cmd := exec.Command("ln", "-s", backupFileName, "latest.tar.gz")
	cmd.Dir = "backups"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	log.Println("Backup completed", backupFilePath)
	return nil
}

func sshIsUp(mine *Mine) bool {
	client, err := sshClient(mine)
	if err != nil {
		return false
	}
	defer client.Close()
	return true
}

func sshClient(m *Mine) (*ssh.Client, error) {
	connStr := fmt.Sprintf("%s:%d", m.PublicIpAddress, m.SSHPort)
	return ssh.DialWithKey(connStr, m.SSHUsername, m.SSHKeyPath)
}

func scpFilesToMine(m *Mine, files []string) error {
	target := fmt.Sprintf("%s@%s:%s", m.SSHUsername, m.PublicIpAddress, "~/")
	for _, file := range files {
		cmd := exec.Command("scp", "-oStrictHostKeyChecking=no", "-i", m.SSHKeyPath, file, target)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func scpFileFromMine(m *Mine, remotePath string, localPath string) error {
	remoteLocation := fmt.Sprintf("%s@%s:%s", m.SSHUsername, m.PublicIpAddress, remotePath)
	cmd := exec.Command("scp", "-oStrictHostKeyChecking=no", "-i", m.SSHKeyPath, remoteLocation, localPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
