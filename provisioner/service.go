package provisioner

import (
	"context"
	"encore.app/pkg/messages"
	provisionedserver "encore.app/provisioner/provisioned_server"
	"encore.dev/rlog"
	"github.com/aneshas/tx/v2"
)

// Store represents a store for provisioned servers
type Store interface {
	Save(ctx context.Context, server provisionedserver.ProvisionedServer) error
	FindExpired(ctx context.Context, limit int) ([]provisionedserver.ProvisionedServer, error)
	Update(ctx context.Context, server provisionedserver.ProvisionedServer) error
}

// Service represents RaceRoom api service
//
// encore:service
type Service struct {
	tx.Transactor

	store Store
	pub   messages.Publisher
}

// HandlePaymentReceived provisions a server after a successful payment
func (s *Service) HandlePaymentReceived(ctx context.Context, msg *messages.ServerPaymentReceived) error {
	return s.WithTransaction(ctx, func(ctx context.Context) error {
		rlog.Info("provisioning server", "id", msg.ServerID)

		// Provision server and fetch instance ID

		instanceID := "some-instance-id"

		err := s.store.Save(
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
			ServerID:   msg.ServerID,
			InstanceID: instanceID,
		})
	})
}

// HandleScheduledTermination terminates server
func (s *Service) HandleScheduledTermination(ctx context.Context, msg *messages.ServerTerminationScheduled) error {
	return s.WithTransaction(ctx, func(ctx context.Context) error {
		rlog.Info("terminating server", "id", msg.ServerID)

		// terminate server

		// update server

		return s.pub.Publish(ctx, messages.ServerTerminated{
			ServerID: msg.ServerID,
		})
	})
}

// ScheduleTermination schedules a batch of expired servers for termination
//
//encore:api private method=GET path=/raceRoom/server/scheduleTermination
func (s *Service) ScheduleTermination(ctx context.Context) error {
	return s.WithTransaction(ctx, func(ctx context.Context) error {
		// TODO - Config
		batch, err := s.store.FindExpired(ctx, 100)
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
