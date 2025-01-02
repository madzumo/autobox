package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// createPEMFile creates a PEM file and saves it locally.
func createPEMFile() (string, error) {
	keyName := "boxPEM"
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", err
	}

	client := ec2.NewFromConfig(cfg)

	resp, err := client.CreateKeyPair(ctx, &ec2.CreateKeyPairInput{
		KeyName: aws.String(keyName),
	})
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("%s.pem", keyName)
	err = os.WriteFile(fileName, []byte(*resp.KeyMaterial), 0600)
	if err != nil {
		return "", err
	}

	fmt.Printf("PEM file created: %s\n", fileName)
	return fileName, nil
}

// createSecurityGroup creates a security group allowing inbound HTTP, HTTPS, and VNC (5901) traffic.
func createSecurityGroup(ctx context.Context, client *ec2.Client, groupName string, description string, vpcID string) (string, error) {
	resp, err := client.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(groupName),
		Description: aws.String(description),
		VpcId:       aws.String(vpcID),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create security group: %v", err)
	}

	securityGroupID := *resp.GroupId

	rules := []struct {
		Protocol string
		Port     int32
	}{
		{"tcp", 80},   // HTTP
		{"tcp", 443},  // HTTPS
		{"tcp", 5901}, // VNC
	}

	for _, rule := range rules {
		_, err := client.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: aws.String(securityGroupID),
			IpPermissions: []types.IpPermission{
				{
					IpProtocol: aws.String(rule.Protocol),
					FromPort:   aws.Int32(rule.Port),
					ToPort:     aws.Int32(rule.Port),
					IpRanges: []types.IpRange{
						{CidrIp: aws.String("0.0.0.0/0")},
					},
				},
			},
		})
		if err != nil {
			return "", fmt.Errorf("failed to add rule to security group: %v", err)
		}
	}

	fmt.Printf("Security group created: %s\n", securityGroupID)
	return securityGroupID, nil
}

// createEC2Instance creates an EC2 instance using the specified PEM file and security group.
func createEC2Instance(ctx context.Context, client *ec2.Client, keyName, securityGroupID, amiID, instanceType string) (string, error) {
	resp, err := client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      aws.String(amiID),
		InstanceType: types.InstanceType(instanceType),
		KeyName:      aws.String(keyName),
		SecurityGroupIds: []string{
			securityGroupID,
		},
		MinCount: aws.Int32(1),
		MaxCount: aws.Int32(1),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create EC2 instance: %v", err)
	}

	instanceID := *resp.Instances[0].InstanceId
	fmt.Printf("EC2 instance created: %s\n", instanceID)
	return instanceID, nil
}
