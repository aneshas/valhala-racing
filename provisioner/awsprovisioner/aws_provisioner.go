package awsprovisioner

import (
	"encore.app/provisioner/provisioned_server"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var serverSpecs = []string{
	"ami-094012549ccc3d684",
	"t2.medium",
	"sg-0b687aa00baccc7ff",
	"raceroom-servers",
}

// NewAWSProvisioner instantiates new aws provisioner
func NewAWSProvisioner(keyID, keySecret, roleARN string) *AWSProvisioner {
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(keyID, keySecret, ""),
		Region:      aws.String(endpoints.EuCentral1RegionID),
	}))

	return &AWSProvisioner{
		ec2: ec2.New(sess, &aws.Config{
			Credentials: stscreds.NewCredentials(sess, roleARN),
		}),
	}
}

// AWSProvisioner represents aws cloud server provisioner
type AWSProvisioner struct {
	ec2 *ec2.EC2
}

// ProvisionServer provisions new server instance
func (p *AWSProvisioner) ProvisionServer(_ uint64) (string, error) {
	result, err := p.ec2.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String(serverSpecs[0]),
		InstanceType: aws.String(serverSpecs[1]),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		SecurityGroupIds: []*string{
			aws.String(serverSpecs[2]),
		},
		KeyName: aws.String(serverSpecs[3]),
	})
	if err != nil {
		return "", err
	}

	id := result.Instances[0].InstanceId

	return *id, nil
}

// TerminateServer terminates server instance
func (p *AWSProvisioner) TerminateServer(instanceID string) error {
	_, err := p.ec2.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	})

	return err
}

// InstanceDetails returns provisioned instance details
func (p *AWSProvisioner) InstanceDetails(instanceID string) (*provisionedserver.InstanceDetails, error) {
	result, err := p.ec2.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	})
	if err != nil {
		return nil, err
	}

	var details provisionedserver.InstanceDetails

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			details.AdminURL = fmt.Sprintf("http://%s:8088", *instance.PublicIpAddress)
			details.IPAddr = *instance.PublicIpAddress
		}
	}

	return &details, nil
}
