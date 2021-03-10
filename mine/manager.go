package mine

import (
	"fmt"

	"b4rt.io/aws-mine-manager/cloud"
)

type MineManager struct {
	cloud      cloud.Cloud
	sshUser    string
	sshPort    int
	sshKeyPath string
	ipID       string
	ipKeeperID string
}

func NewManager(cloud cloud.Cloud, sshUser string, sshPort int, sshKeyPath string, ipID string, ipKeeperID string) (MineManager, error) {
	return MineManager{
		cloud:      cloud,
		sshUser:    sshUser,
		sshPort:    sshPort,
		sshKeyPath: sshKeyPath,
		ipID:       ipID,
		ipKeeperID: ipKeeperID,
	}, nil
}

func (m MineManager) GetMines() ([]Mine, error) {
	instances, err := m.cloud.GetInstances()
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
	i, err := m.cloud.CreateInstance(imageId, keyName, secGroups)
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

func (m MineManager) GetAddresses() {
	m.cloud.GetAddresses()
}

func (m MineManager) Terminate(mineID string) error {
	return m.cloud.Terminate(mineID)
}

func (m *MineManager) AssignElasticIP(mineID string) error {
	return m.cloud.AssignIP(m.ipID, mineID)
}

// Switch ElasticIP to dummy instance(free). ElasticIP are chareged if not associated with instance
func (m MineManager) ReleaseElasticIP() error {
	return m.cloud.AssignIP(m.ipID, m.ipKeeperID)
}
