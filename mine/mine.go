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

func (m *Mine) IsReady() bool {
	if m.State != "running" {
		return false
	}
	if !sshIsUp(m) {
		return false
	}
	return true
}

func (m *Mine) Setup(setupPath string, setupFiles []string, setupCmds []string, backupsPath string, latestBackupName string, dataDirPath string) error {
	log.Println("Setup new Mine")
	log.Println("Copying setup files to mine")

	for _, setupFile := range setupFiles {
		err := scpToMine(m, fmt.Sprintf("%s%s", setupPath, setupFile))
		if err != nil {
			return err
		}
	}

	log.Println("Copying latest data to mine")
	err := scpToMine(m, fmt.Sprintf("%s/%s", backupsPath, latestBackupName))
	if err != nil {
		return err
	}

	client, err := sshClient(m)
	if err != nil {
		return err
	}
	defer client.Close()

	log.Println("Extracting latest data")
	extractCmd := fmt.Sprintf("mkdir -p %s && tar -C %s -zxf %s", dataDirPath, dataDirPath, latestBackupName)
	err = execShell(client, extractCmd)
	if err != nil {
		return fmt.Errorf("error taring data %s", err.Error())
	}

	for _, setupCmd := range setupCmds {
		log.Println("Running setup command:", setupCmd)
		output, err := client.Cmd(setupCmd).SmartOutput()
		if err != nil {
			log.Println(string(output))
			return err
		}
	}
	return nil
}

func (m *Mine) Players() error {
	client, err := sshClient(m)
	if err != nil {
		return err
	}
	defer client.Close()

	cmd := "sudo docker exec minecraft-server rcon-cli list"
	out, err := client.Cmd(cmd).SmartOutput()
	log.Println(string(out))
	return err
}

func (m *Mine) Save() error {
	client, err := sshClient(m)
	if err != nil {
		return err
	}
	defer client.Close()

	cmds := []string{
		"sudo docker exec minecraft-server rcon-cli save-all",
		// "sudo docker exec minecraft-server rcon-cli stop",
	}

	for _, cmd := range cmds {
		out, err := client.Cmd(cmd).SmartOutput()
		log.Println(string(out))
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Mine) Backup(dataPath string, backupDirPath string, backupFileFmt string, latestDataFileName string) error {
	client, err := sshClient(m)
	if err != nil {
		return err
	}
	defer client.Close()

	remoteArchivePath := "~/data.tar.gz"
	bckpCmd := fmt.Sprintf("cd %s && tar czf %s .", dataPath, remoteArchivePath)
	err = execShell(client, bckpCmd)
	if err != nil {
		return fmt.Errorf("error taring data %s", err.Error())
	}

	now := time.Now()
	backupFileName := fmt.Sprintf(
		backupFileFmt, now.Day(), now.Month(), now.Year(), now.Hour(), now.Minute())
	backupFilePath := fmt.Sprintf("%s/%s", backupDirPath, backupFileName)

	err = scpFileFromMine(m, remoteArchivePath, backupFilePath)
	if err != nil {
		return err
	}

	cmd := exec.Command("ln", "-sF", backupFileName, latestDataFileName)
	cmd.Dir = backupDirPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	log.Println("Backup completed", backupFilePath)
	return nil
}

func execShell(client *ssh.Client, cmd string) error {
	out, err := client.Cmd(cmd).SmartOutput()
	log.Println(string(out))
	if err != nil {
		log.Println(string(out))
	}
	return err
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

func scpToMine(m *Mine, file string) error {
	target := fmt.Sprintf("%s@%s:%s", m.SSHUsername, m.PublicIpAddress, "~/")
	// log.Println("scp", "-oStrictHostKeyChecking=no", "-i", m.SSHKeyPath, file, target)
	cmd := exec.Command("scp", "-oStrictHostKeyChecking=no", "-i", m.SSHKeyPath, file, target)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
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
