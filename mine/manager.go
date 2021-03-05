package mine

import (
	"log"

	"b4rt.io/aws-mine-manager/cloud"
	ssh "github.com/helloyi/go-sshclient"
)

type MineManager struct {
	Cloud cloud.Cloud
}

type MineInstance interface {
}

type Mine struct {
	Id              string
	PublicIpAddress string
	State           string
}

func (m *Mine) Connect() {
	client, err := ssh.DialWithKey(m.PublicIpAddress+":22", "ubuntu", "/Users/bart/poligon/aws/keys/awsJumbox.pem")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	out, err := client.Cmd("pwd; ls -la ;docker -v").Output()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(out))
}

func NewManager(cloud cloud.Cloud) (MineManager, error) {
	log.Println("Create mine manager")
	return MineManager{
		Cloud: cloud,
	}, nil
}

func (m MineManager) GetMines() ([]Mine, error) {
	instances, err := m.Cloud.GetInstances()
	if err != nil {
		return []Mine{}, err
	}

	var mines []Mine
	for _, i := range instances {
		if i.State == "terminated" {
			continue
		}

		mineInstance := false
		if val, ok := i.Tags["mine-node"]; ok {
			if val == "true" {
				mineInstance = true
			}
		}

		if !mineInstance {
			continue
		}

		mines = append(mines, Mine{
			Id:              i.Id,
			PublicIpAddress: i.PublicIP,
			State:           i.State,
		})
	}
	return mines, nil
}

func (m MineManager) CreateMine(imageId string, keyName string, secGroups string) {
	log.Println("create new mine")
	m.Cloud.CreateInstance(imageId, keyName, secGroups)
}
