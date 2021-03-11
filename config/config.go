package config

import "github.com/joeshaw/envdecode"

type Config struct {
	AWS struct {
		ImageId       string `env:"AWS_IMAGE_ID,default=ami-0767046d1677be5a0"`
		SecurityGroup string `env:"AWS_SEC_GROUP,required"`
		Key           struct {
			Name string `env:"AWS_KEY_NAME,required"`
			Path string `env:"AWS_KEY_PATH,required"`
		}
		SSH struct {
			User string `env:"AWS_SSH_USERNAME,default=ubuntu"`
			Port int    `env:"AWS_SSH_PORT,default=22"`
		}
		ElasticIP struct {
			AllocID                 string `env:"AWS_ELASTIC_IP_ID,required"`
			ReleaseToNetInterfaceID string `env:"AWS_ELASTIC_IP_RELEASE_TO_NET_INT_ID,required"`
		}
	}
	Mine struct {
		Name               string `env:"MINE_INSTANCE_NAME,default=Minecraft poligon dzieciakow"`
		DataDirPath        string `env:"MINE_DATA_DIR_PATH,default=data/"`
		BackupDirPath      string `env:"MINE_BACKUP_DIR_PATH,default=backups/"`
		BackupFileFormat   string `env:"MINE_BACKUP_FILE_FORMAT,default=data_%02d-%02d-%dT%02d%02d.tar.gz"`
		SetupDirPath       string `env:"MINE_SETUP_DIR_PATH,default=setup/"`
		SetupFiles         string `env:"MINE_SETUP_FILES,default=setup.sh run.sh"`
		SetupCmds          string `env:"MINE_SETUP_CMDS,default=./setup.sh ./run.sh"`
		LatestDataFileName string `env:"MINE_LATEST_DATA_FILE_NAME,default=latest.tar.gz"`
	}
}

func LoadFromEnv() (config Config, err error) {
	err = envdecode.StrictDecode(&config)
	if err != nil {
		return
	}
	return
}
