package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type AWS struct {
	Region         string `json:"region"`
	PemKeyFileName string `json:"pemkeyfilename"`
	AmiID          string `json:"amiid"`
	InstanceType   string `json:"instancetype"`
	Key            string `json:"key"`
	Secret         string `json:"secret"`
}
type EC2InstanceIP struct {
	InstanceID string
	PublicIP   string
	PrivateIP  string
}

func (a *AWS) createEc2Client() (*ec2.Client, error) {
	ctx := context.Background()
	customCreds := aws.NewCredentialsCache(
		credentials.NewStaticCredentialsProvider(a.Key, a.Secret, ""),
	)
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(customCreds), config.WithRegion(a.Region))
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	return client, nil
}

func (a *AWS) getActiveEC2s(client *ec2.Client) (int, error) {
	ctx := context.Background()

	// Describe instances with the AUTO-BOX tag
	resp, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:AUTO-BOX"),
				Values: []string{"true"},
			},
			{
				Name:   aws.String("instance-state-code"),
				Values: []string{"16"},
			},
		},
	})

	if err != nil {
		return 0, err
	}

	return len(resp.Reservations), nil
}

func (a *AWS) createPEMFile(client *ec2.Client) error {
	ctx := context.Background()
	// Check if the key pair already exists
	existingKEY, err := client.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("key-name"),
				Values: []string{a.PemKeyFileName},
			},
		},
	})
	if err != nil {
		return err
	}
	// If a security group with the given name exists, return its ID
	if len(existingKEY.KeyPairs) > 0 {
		return nil
	}

	resp, err := client.CreateKeyPair(ctx, &ec2.CreateKeyPairInput{
		KeyName: aws.String(a.PemKeyFileName),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeKeyPair,
				Tags: []types.Tag{
					{
						Key:   aws.String("AUTO-BOX"),
						Value: aws.String("true"),
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	// Save PEM key in current directory instead of subfolder
	currentDir, _ := os.Getwd()
	fmt.Printf("DEBUG createPEMFile: Saving PEM key in current directory: %s\n", currentDir)

	// fileName := fmt.Sprintf("%s.pem", keyName)
	fileName := filepath.Join(currentDir, fmt.Sprintf("%s.pem", a.PemKeyFileName))
	fmt.Printf("DEBUG createPEMFile: PEM file path: %s\n", fileName)
	err = os.WriteFile(fileName, []byte(*resp.KeyMaterial), 0400)
	if err != nil {
		return err
	}

	a.restrictWindowsFilePermissions(fileName)
	return nil
}

// will install in default VPC
func (a *AWS) createSecurityGroup(sgName, description string, client *ec2.Client) (string, error) {
	ctx := context.Background()

	existingGroups, err := client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("group-name"),
				Values: []string{sgName},
			},
		},
	})
	if err != nil {
		return "", err
	}

	// If a security group with the given name exists, return its ID
	if len(existingGroups.SecurityGroups) > 0 {
		return *existingGroups.SecurityGroups[0].GroupId, nil
	}

	resp, err := client.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(sgName),
		Description: aws.String(description),
		// VpcId:       aws.String(vpcID),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeSecurityGroup,
				Tags: []types.Tag{
					{
						Key:   aws.String("AUTO-BOX"),
						Value: aws.String("true"),
					},
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	securityGroupID := *resp.GroupId

	rules := []struct {
		Protocol string
		Port     int32
	}{
		{"tcp", 80},   // HTTP
		{"tcp", 443},  // HTTPS
		{"tcp", 5901}, // VNC
		{"tcp", 22},   //telnet
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
			return "", err
		}
	}

	// fmt.Printf("Security group created: %s\n", securityGroupID)
	return securityGroupID, nil
}

func (a *AWS) validateAMI(client *ec2.Client) error {
	ctx := context.Background()

	// Check if the current AMI exists in this region
	resp, err := client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		ImageIds: []string{a.AmiID},
	})
	if err != nil {
		fmt.Printf("Error checking AMI %s in region %s: %v\n", a.AmiID, a.Region, err)
	} else if len(resp.Images) > 0 {
		fmt.Printf("AMI %s found and valid in region %s\n", a.AmiID, a.Region)
		return nil // AMI is valid, use it
	}

	fmt.Printf("AMI %s not found in region %s, searching for alternative Ubuntu AMIs...\n", a.AmiID, a.Region)

	// Try multiple Ubuntu versions in order of preference
	searchPatterns := []string{
		"ubuntu/images/hvm-ssd/ubuntu-noble-24.04-amd64-server-*", // Ubuntu 24.04 LTS
		"ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*", // Ubuntu 22.04 LTS
		"ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*", // Ubuntu 20.04 LTS
	}

	for _, pattern := range searchPatterns {
		fmt.Printf("Searching for pattern: %s\n", pattern)

		resp, err = client.DescribeImages(ctx, &ec2.DescribeImagesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("name"),
					Values: []string{pattern},
				},
				{
					Name:   aws.String("owner-id"),
					Values: []string{"099720109477"}, // Canonical's AWS account ID
				},
				{
					Name:   aws.String("state"),
					Values: []string{"available"},
				},
			},
		})
		if err != nil {
			fmt.Printf("Error searching for pattern %s: %v\n", pattern, err)
			continue
		}

		if len(resp.Images) > 0 {
			// Use the most recent AMI
			latestAMI := resp.Images[0]
			for _, img := range resp.Images {
				if img.CreationDate != nil && latestAMI.CreationDate != nil {
					if *img.CreationDate > *latestAMI.CreationDate {
						latestAMI = img
					}
				}
			}

			a.AmiID = *latestAMI.ImageId
			fmt.Printf("Found and using AMI: %s (%s)\n", a.AmiID, *latestAMI.Name)
			return nil
		}
	}

	return fmt.Errorf("no suitable Ubuntu AMI found in region %s. Please check your region or specify a valid AMI ID for %s", a.Region, a.Region)
}

func (a *AWS) createEC2Instance(securityGroupID string, client *ec2.Client, batchT string) error {
	ctx := context.Background()

	// Debug: Print the parameters being used
	fmt.Printf("Creating EC2 with parameters:\n")
	fmt.Printf("  AMI ID: %s\n", a.AmiID)
	fmt.Printf("  Instance Type: %s\n", a.InstanceType)
	fmt.Printf("  Key Name: %s\n", a.PemKeyFileName)
	fmt.Printf("  Security Group ID: %s\n", securityGroupID)
	fmt.Printf("  Region: %s\n", a.Region)

	// First try with spot instances
	runInput := &ec2.RunInstancesInput{
		ImageId:      aws.String(a.AmiID),
		InstanceType: types.InstanceType(a.InstanceType),
		KeyName:      aws.String(a.PemKeyFileName),
		SecurityGroupIds: []string{
			securityGroupID,
		},
		MinCount: aws.Int32(1),
		MaxCount: aws.Int32(1),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags: []types.Tag{
					{
						Key:   aws.String("AUTO-BOX"),
						Value: aws.String("true"),
					},
					{
						Key:   aws.String("BatchTag"),
						Value: aws.String(batchT),
					},
				},
			},
		},
		InstanceMarketOptions: &types.InstanceMarketOptionsRequest{
			MarketType: types.MarketTypeSpot,
		},
	}

	resp, err := client.RunInstances(ctx, runInput)
	if err != nil {
		// If spot instance fails, try regular on-demand instance
		fmt.Printf("Spot instance failed, trying on-demand: %v\n", err)

		runInput.InstanceMarketOptions = nil // Remove spot instance option
		resp, err = client.RunInstances(ctx, runInput)
		if err != nil {
			return fmt.Errorf("failed to create EC2 instance (both spot and on-demand): %w", err)
		}
	}

	instanceID := *resp.Instances[0].InstanceId
	fmt.Printf("EC2 instance created: %s\n", instanceID)
	return nil
}

func (a *AWS) deleteEC2Instances(client *ec2.Client, batchT string) error {
	ctx := context.Background()

	// Describe instances with the AUTO-BOX tag
	resp, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:AUTO-BOX"),
				Values: []string{"true"},
			},
		},
	})
	if batchT != "" {
		resp, err = client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("tag:AUTO-BOX"),
					Values: []string{"true"},
				},
				{
					Name:   aws.String("tag:BatchTag"),
					Values: []string{batchT},
				},
			},
		})
	}
	if err != nil {
		return err
	}

	var instanceIDs []string
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			instanceIDs = append(instanceIDs, *instance.InstanceId)
		}
	}

	if len(instanceIDs) == 0 {
		fmt.Println("No EC2 instances with tag AUTO-BOX found.")
		return nil
	}

	// Terminate the instances
	_, err = client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: instanceIDs,
	})
	if err != nil {
		return err
	}

	// //wait for them to terminate fully
	// var allAreTerminated bool
	// for {
	// 	resp, err = client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
	// 		Filters: []types.Filter{
	// 			{
	// 				Name:   aws.String("tag:AUTO-BOX"),
	// 				Values: []string{"true"},
	// 			},
	// 		},
	// 	})
	// 	if err != nil {
	// 		return err
	// 	}
	// 	allAreTerminated = true
	// 	for _, reservation := range resp.Reservations {
	// 		for _, instance := range reservation.Instances {
	// 			for {
	// 				if *instance.State.Code != 48 { //0-pending, 16-running, 32-shutting-down, 48-terminated
	// 					allAreTerminated = false
	// 					fmt.Printf("name: %s state: %d\n", *instance.InstanceId, *instance.State.Code)
	// 				}
	// 			}

	// 		}
	// 	}
	// 	if allAreTerminated {
	// 		break
	// 	} else {
	// 		time.Sleep(10 * time.Second)
	// 	}
	// }

	fmt.Printf("Terminated All EC2 instances: %v\n", instanceIDs)
	return nil
}

// func (a *AWS) deleteSecurityGroups(client *ec2.Client) error {
// 	ctx := context.Background()

// 	// Describe security groups with the AUTO-BOX tag
// 	resp, err := client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
// 		Filters: []types.Filter{
// 			{
// 				Name:   aws.String("tag:AUTO-BOX"),
// 				Values: []string{"true"},
// 			},
// 		},
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	for _, sg := range resp.SecurityGroups {
// 		_, err := client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
// 			GroupId: sg.GroupId,
// 		})
// 		if err != nil {
// 			fmt.Printf("Failed to delete security group %s: %v\n", *sg.GroupId, err)
// 		} else {
// 			fmt.Printf("Deleted security group: %s\n", *sg.GroupId)
// 		}
// 	}

// 	return nil
// }

func (a *AWS) deletePEMFile(client *ec2.Client) error {
	ctx := context.Background()

	// Delete the key pair
	_, err := client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{
		KeyName: aws.String(a.PemKeyFileName),
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *AWS) compileIPaddressesAws(client *ec2.Client, batchT string) (ips []string, fullEC2 []EC2InstanceIP, err error) {
	ctx := context.Background()

	// Describe EC2 instances
	resp, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:AUTO-BOX"),
				Values: []string{"true"},
			},
			{
				Name:   aws.String("tag:BatchTag"),
				Values: []string{batchT},
			},
		},
	})
	if err != nil {
		return nil, nil, err
	}
	if batchT == "" {
		resp, err = client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("tag:AUTO-BOX"),
					Values: []string{"true"},
				},
			},
		})
		if err != nil {
			return nil, nil, err
		}
	}

	// Iterate over reservations and instances to collect IP addresses
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			ipInfo := EC2InstanceIP{
				InstanceID: *instance.InstanceId,
			}

			// Check if Public IP exists
			if instance.PublicIpAddress != nil {
				ipInfo.PublicIP = *instance.PublicIpAddress
				ips = append(ips, *instance.PublicIpAddress)
			}

			// Check if Private IP exists
			if instance.PrivateIpAddress != nil {
				ipInfo.PrivateIP = *instance.PrivateIpAddress
			}

			fullEC2 = append(fullEC2, ipInfo)
		}
	}

	return ips, fullEC2, nil
}

func (a *AWS) restrictWindowsFilePermissions(fileName string) error {
	fmt.Printf("DEBUG restrictWindowsFilePermissions: Setting SSH-compatible permissions on: %s\n", fileName)

	// Use icacls command to set proper Windows permissions for SSH private keys
	// This removes all permissions except for the current user
	cmd := exec.Command("icacls", fileName, "/inheritance:r", "/grant:r", fmt.Sprintf("%s:(R)", os.Getenv("USERNAME")))

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("DEBUG restrictWindowsFilePermissions: icacls command failed: %v\n", err)
		fmt.Printf("DEBUG restrictWindowsFilePermissions: icacls output: %s\n", string(output))

		// Fallback: try using attrib command to at least make it read-only
		fmt.Printf("DEBUG restrictWindowsFilePermissions: Falling back to attrib command\n")
		cmd2 := exec.Command("attrib", "+R", fileName)
		if err2 := cmd2.Run(); err2 != nil {
			fmt.Printf("DEBUG restrictWindowsFilePermissions: attrib command also failed: %v\n", err2)
			return fmt.Errorf("failed to set file permissions: icacls failed with %v, attrib failed with %v", err, err2)
		}
		return nil
	}

	fmt.Printf("DEBUG restrictWindowsFilePermissions: Successfully set SSH-compatible permissions\n")
	fmt.Printf("DEBUG restrictWindowsFilePermissions: icacls output: %s\n", string(output))

	return nil
}

// func ec2StatusReady () bool {

// }
