package mine

import (
	"fmt"

	"b4rt.io/aws-mine-manager/cloud"
)

type MineManager struct {
	Cloud      cloud.Cloud
	sshUser    string
	sshPort    int
	sshKeyPath string
}

type MineInstance interface {
}

func NewManager(cloud cloud.Cloud, sshUser string, sshPort int, sshKeyPath string) (MineManager, error) {
	return MineManager{
		Cloud:      cloud,
		sshUser:    sshUser,
		sshPort:    sshPort,
		sshKeyPath: sshKeyPath,
	}, nil
}

func (m MineManager) GetMines() ([]Mine, error) {
	instances, err := m.Cloud.GetInstances()
	if err != nil {
		return []Mine{}, err
	}

	var mines []Mine
	for _, i := range instances {
		if i.State == "terminated" || i.State == "shutting-down" {
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
			SSHUsername:     m.sshUser,
			SSHPort:         m.sshPort,
			SSHKeyPath:      m.sshKeyPath,
		})
	}
	return mines, nil
}

func (m MineManager) GetMine(mineId string) (Mine, error) {
	mines, err := m.GetMines()
	if err != nil {
		return Mine{}, err
	}
	for _, mine := range mines {
		if mine.Id == mineId {
			return mine, nil
		}
	}
	return Mine{}, fmt.Errorf("Mine with id:%s not found", mineId)
}

func (m MineManager) CreateMine(imageId string, keyName string, secGroups string) (Mine, error) {
	i, err := m.Cloud.CreateInstance(imageId, keyName, secGroups)
	if err != nil {
		return Mine{}, err
	}

	return Mine{
		Id:              i.Id,
		PublicIpAddress: i.PublicIP,
		State:           i.State,
		SSHUsername:     m.sshUser,
		SSHPort:         m.sshPort,
		SSHKeyPath:      m.sshKeyPath,
	}, err
}
