package config

import "github.com/joeshaw/envdecode"

type Config struct {
	AWSImageId           string `env:"AWS_IMAGE_ID,default=ami-0767046d1677be5a0"`
	AWSSecurityGroup     string `env:"AWS_SEC_GROUP,required"`
	AWSKeyName           string `env:"AWS_KEY_NAME,required"`
	AWSKeyPath           string `env:"AWS_KEY_PATH,required"`
	AWSSSHUser           string `env:"AWS_SSH_USERNAME,default=ubuntu"`
	AWSSSSPort           int    `env:"AWS_SSH_PORT,default=22"`
	AWSElasticIPID       string `env:"AWS_ELASTIC_IP_ID,required"`
	AWSElasticIPKeeperID string `env:"AWS_ELASTIC_IP_KEEPER_ID,required"`
}

func LoadFromEnv() (config Config, err error) {
	err = envdecode.StrictDecode(&config)
	if err != nil {
		return
	}
	return
}
