package emailer

import (
	"context"
	"encore.app/pkg/infra"
	"encore.app/pkg/messages"
	"encore.app/raceroom"
	"encore.dev/pubsub"
	"encore.dev/rlog"
)

var _ = pubsub.NewSubscription(
	infra.ProvisionedServersTopic,
	"provisioned-server-email-notifier",
	pubsub.SubscriptionConfig[*messages.ServerProvisioned]{
		Handler: func(ctx context.Context, msg *messages.ServerProvisioned) error {
			rlog.Info("sending email")

			_, _ = raceroom.ServerDetails(ctx, msg.ServerID)

			// TODO - CC me - configurable

			return nil
		},
		MaxConcurrency: 20,
		RetryPolicy: &pubsub.RetryPolicy{
			MaxRetries: 20,
		},
	},
)

var _ = pubsub.NewSubscription(
	infra.TerminatedServersTopic,
	"provisioned-server-email-notifier",
	pubsub.SubscriptionConfig[*messages.ServerTerminated]{
		Handler: func(ctx context.Context, msg *messages.ServerTerminated) error {
			rlog.Info("sending email")

			_, _ = raceroom.ServerDetails(ctx, msg.ServerID)

			// TODO - CC me - configurable

			return nil
		},
		MaxConcurrency: 20,
		RetryPolicy: &pubsub.RetryPolicy{
			MaxRetries: 20,
		},
	},
)
