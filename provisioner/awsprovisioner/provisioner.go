package awsprovisioner

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

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
func (p *AWSProvisioner) ProvisionServer(serverID uint64) (string, error) {
	result, err := p.ec2.RunInstances(&ec2.RunInstancesInput{
		DryRun:       aws.Bool(true),
		ImageId:      aws.String("ami-094012549ccc3d684"),
		InstanceType: aws.String("t2.medium"),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		SecurityGroupIds: []*string{
			aws.String("sg-0b687aa00baccc7ff"),
		},
		KeyName: aws.String("raceroom-servers"),
	})
	if err != nil {
		return "", err
	}

	id := result.Instances[0].InstanceId

	return *id, nil
}

// TerminateServer terminates server instance
func (p *AWSProvisioner) TerminateServer(instanceID string) error {
	return nil
}
