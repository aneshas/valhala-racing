package awsprovisioner

import (
	provisionedserver "encore.app/provisioner/provisioned_server"
	"fmt"
)

// TestProvisioner is an instance of noop provisioner
var TestProvisioner = NoOpProvisioner{}

// NoOpProvisioner represents dummy provisioner
type NoOpProvisioner struct{}

// ProvisionServer fakes provisioning
func (n NoOpProvisioner) ProvisionServer(serverID uint64) (string, error) {
	return fmt.Sprintf("dummy-instance-id-%d", serverID), nil
}

// TerminateServer fakes termination
func (n NoOpProvisioner) TerminateServer(_ string) error {
	return nil
}

// InstanceDetails returns dummy instance details
func (n NoOpProvisioner) InstanceDetails(instanceID string) (*provisionedserver.InstanceDetails, error) {
	return &provisionedserver.InstanceDetails{
		AdminURL: fmt.Sprintf("http://dummy-server-%s.com", instanceID),
		IPAddr:   "10.11.12.13",
	}, nil
}
