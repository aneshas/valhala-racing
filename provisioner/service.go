package provisioner

import (
	"context"
	"encore.app/pkg/messages"
	provisionedserver "encore.app/provisioner/provisioned_server"
	"encore.dev/rlog"
	"github.com/aneshas/tx/v2"
	"time"
)

// Store represents a store for provisioned servers
type Store interface {
	Save(ctx context.Context, server provisionedserver.ProvisionedServer) (uint64, error)
	FindByID(ctx context.Context, id uint64) (*provisionedserver.ProvisionedServer, error)
	FindExpired(ctx context.Context, limit int) ([]provisionedserver.ProvisionedServer, error)
	Update(ctx context.Context, server provisionedserver.ProvisionedServer) error
}

// ServerProvisioner represents a cloud / on-prem server provisioner
type ServerProvisioner interface {
	ProvisionServer(serverID uint64) (string, error)
	TerminateServer(instanceID string) error
}

// Service represents provisioner api service
//
// encore:service
type Service struct {
	tx.Transactor

	provisioner ServerProvisioner
	store       Store
	pub         messages.Publisher
}

// InstanceDetailsResp represents server instance information
type InstanceDetailsResp struct {
	AdminURL  string `json:"adminUrl,omitempty"`
	IPAddr    string `json:"IPAddr,omitempty"`
	ExpiresOn string `json:"expiryTime"`
}

//encore:api private method=GET path=/provisioner/instance/:id/details
func (s *Service) InstanceDetails(ctx context.Context, id string) (*InstanceDetailsResp, error) {
	return &InstanceDetailsResp{
		AdminURL:  "https:/foo-bar.com",
		IPAddr:    "10.18.2.3",
		ExpiresOn: time.Now().Format(time.RFC1123), // From DB
	}, nil
}

// HandlePaymentReceived provisions a server after a successful payment
//
//encore:api private method=POST path=/provisioner/handlePaymentReceived
func (s *Service) HandlePaymentReceived(ctx context.Context, msg *messages.ServerPaymentReceived) error {
	return s.WithTransaction(ctx, func(ctx context.Context) error {
		instanceID, err := s.provisioner.ProvisionServer(msg.ServerID)
		if err != nil {
			rlog.Error("provisioner error", "err", err.Error())
			instanceID = "xxx"

			return nil
		}

		provisionedServerID, err := s.store.Save(
			ctx,
			*provisionedserver.New(
				msg.ServerID,
				instanceID,
				msg.HoursReserved,
			),
		)
		if err != nil {
			return err
		}

		return s.pub.Publish(ctx, messages.ServerProvisioned{
			ServerID:   provisionedServerID,
			InstanceID: instanceID,
		})
	})
}

// HandleScheduledTermination terminates server
//
//encore:api private method=POST path=/provisioner/terminate
func (s *Service) HandleScheduledTermination(ctx context.Context, msg *messages.ServerTerminationScheduled) error {
	return s.WithTransaction(ctx, func(ctx context.Context) error {
		server, err := s.store.FindByID(ctx, msg.ServerID)
		if err != nil {
			return err
		}

		server.Terminate()

		err = s.provisioner.TerminateServer(server.InstanceID)
		if err != nil {
			return err
		}

		err = s.store.Update(ctx, *server)
		if err != nil {
			return err
		}

		return s.pub.Publish(ctx, messages.ServerTerminated{
			ServerID: msg.ServerID,
		})
	})
}

const terminationBatchSize = 20

// ScheduleTermination schedules a batch of expired servers for termination
//
//encore:api private method=GET path=/provisioner/scheduleTermination
func (s *Service) ScheduleTermination(ctx context.Context) error {
	return s.WithTransaction(ctx, func(ctx context.Context) error {
		batch, err := s.store.FindExpired(ctx, terminationBatchSize)
		if err != nil {
			return err
		}

		for i := range batch {
			svr := &batch[i]

			svr.ScheduleTermination()

			err = s.store.Update(ctx, *svr)
			if err != nil {
				return err
			}

			err = s.pub.Publish(ctx, messages.ServerTerminationScheduled{
				ServerID:   svr.ID,
				InstanceID: svr.InstanceID,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
}
