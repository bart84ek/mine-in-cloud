package cloud

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type AWSCloud struct {
	ec2    *ec2.Client
	config aws.Config
	ctx    context.Context
}

func AWS() (AWSCloud, error) {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return AWSCloud{}, err
	}

	ec2client := ec2.NewFromConfig(cfg)
	return AWSCloud{
		ec2:    ec2client,
		config: cfg,
		ctx:    ctx,
	}, nil
}

func (c AWSCloud) GetInstances() ([]Instance, error) {
	input := &ec2.DescribeInstancesInput{}
	results, err := c.ec2.DescribeInstances(c.ctx, input)
	if err != nil {
		return []Instance{}, err
	}

	var instances []Instance
	for _, r := range results.Reservations {
		for _, i := range r.Instances {
			tags := make(map[string]string)
			for _, tag := range i.Tags {
				tags[*tag.Key] = *tag.Value
			}
			publicIP := "0.0.0.0"
			if i.PublicIpAddress != nil {
				publicIP = *i.PublicIpAddress
			}
			instances = append(instances, Instance{
				Id:            *i.InstanceId,
				ReservationId: *r.ReservationId,
				PublicIP:      publicIP,
				State:         string(i.State.Name),
				Tags:          tags,
			})
		}

	}
	return instances, err
}

func (c AWSCloud) CreateInstance(imageId string, keyName string, secGroup string) (Instance, error) {
	result, err := c.ec2.RunInstances(c.ctx, &ec2.RunInstancesInput{
		ImageId:  &imageId,
		MinCount: 1,
		MaxCount: 1,
		// InstanceType: types.InstanceTypeT2Micro,
		InstanceType:   types.InstanceTypeT2Medium,
		KeyName:        &keyName,
		SecurityGroups: []string{secGroup},
	})
	if err != nil {
		return Instance{}, err
	}

	_, err = c.ec2.CreateTags(c.ctx, &ec2.CreateTagsInput{
		Resources: []string{*result.Instances[0].InstanceId},
		Tags: []types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("Minecraft Poligon dzieciakow"),
			},
			{
				Key:   aws.String("mine-node"),
				Value: aws.String("true"),
			},
		},
	})
	if err != nil {
		return Instance{}, err
	}

	newInstance := result.Instances[0]

	tags := make(map[string]string)
	tags["mine-node"] = "true"

	publicIP := "0.0.0.0"
	if newInstance.PublicIpAddress != nil {
		publicIP = *newInstance.PublicIpAddress
	}

	return Instance{
		Id:            *newInstance.InstanceId,
		ReservationId: *result.ReservationId,
		PublicIP:      publicIP,
		State:         string(newInstance.State.Name),
		Tags:          tags,
	}, nil
}
