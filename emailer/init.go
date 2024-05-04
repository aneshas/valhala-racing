package emailer

import (
	"encore.app/pkg/infra"
	"encore.app/pkg/messages"
	"encore.dev/pubsub"
	brevo "github.com/getbrevo/brevo-go/lib"
)

var secrets struct {
	BrevoAPIKey string
}

func initService() (*Service, error) {
	cfg := brevo.NewConfiguration()

	cfg.AddDefaultHeader("api-key", secrets.BrevoAPIKey)

	return &Service{
		client: brevo.NewAPIClient(cfg),
	}, nil
}

var _ = pubsub.NewSubscription(
	infra.ProvisionedServersTopic,
	"provisioned-server-email-notifier",
	pubsub.SubscriptionConfig[*messages.ServerProvisioned]{
		Handler:        pubsub.MethodHandler((*Service).HandleServerProvisioned),
		MaxConcurrency: 20,
		RetryPolicy: &pubsub.RetryPolicy{
			MaxRetries: 20,
		},
	},
)

var _ = pubsub.NewSubscription(
	infra.TerminatedServersTopic,
	"terminated-server-email-notifier",
	pubsub.SubscriptionConfig[*messages.ServerTerminated]{
		Handler:        pubsub.MethodHandler((*Service).HandleServerTerminated),
		MaxConcurrency: 20,
		RetryPolicy: &pubsub.RetryPolicy{
			MaxRetries: 20,
		},
	},
)
