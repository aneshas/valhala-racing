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
	InstanceDetails(instanceID string) (*provisionedserver.InstanceDetails, error)
}

// DeDupFunc represents message deduplication func
type DeDupFunc func(ctx context.Context, key string) (bool, error)

// Service represents provisioner api service
//
// encore:service
type Service struct {
	tx.Transactor

	provisioner      ServerProvisioner
	store            Store
	isMessageSeenFor DeDupFunc
	pub              messages.Publisher
}

// InstanceDetailsResp represents server instance information
type InstanceDetailsResp struct {
	AdminURL  string `json:"adminUrl,omitempty"`
	IPAddr    string `json:"IPAddr,omitempty"`
	ExpiresOn string `json:"expiryTime"`
}

//encore:api private method=GET path=/provisioner/instance/:provisionedServerId/details
func (s *Service) InstanceDetails(ctx context.Context, provisionedServerId uint64) (*InstanceDetailsResp, error) {
	server, err := s.store.FindByID(ctx, provisionedServerId)
	if err != nil {
		return nil, err
	}

	details, err := s.provisioner.InstanceDetails(server.InstanceID)
	if err != nil {
		return nil, err
	}

	return &InstanceDetailsResp{
		AdminURL:  details.AdminURL,
		IPAddr:    details.IPAddr,
		ExpiresOn: server.ExpiresAt.Format(time.RFC1123),
	}, nil
}

// HandlePaymentReceived provisions a server after a successful payment
//
//encore:api private method=POST path=/provisioner/handlePaymentReceived
func (s *Service) HandlePaymentReceived(ctx context.Context, msg *messages.ServerPaymentReceived) error {
	// Tried implementing this as a middleware but middleware would not register
	// to a pubsub service handler method such as this one for some reason...
	ok, err := s.isMessageSeenFor(ctx, msg.CacheKey())
	if err != nil {
		return err
	}

	if ok {
		return nil
	}

	return s.WithTransaction(ctx, func(ctx context.Context) error {
		instanceID, err := s.provisioner.ProvisionServer(msg.ServerID)
		if err != nil {
			rlog.Error("provisioner error", "err", err.Error())

			return err
		}

		id, err := s.store.Save(
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
			ServerID:            msg.ServerID,
			ProvisionedServerID: id,
			InstanceID:          instanceID,
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
			ServerID: server.ServerID,
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
