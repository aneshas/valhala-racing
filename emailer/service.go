package emailer

import (
	"context"
	"encore.app/checkout"
	"encore.app/pkg/infra"
	"encore.app/pkg/messages"
	"encore.dev/pubsub"
	"encore.dev/rlog"
	brevo "github.com/getbrevo/brevo-go/lib"
)

var secrets struct {
	// AWSKeyID     string
	// AWSKeySecret string
	// AWSRoleARN   string
	BrevoAPIKey string
}

var _ = pubsub.NewSubscription(
	infra.ProvisionedServersTopic,
	"provisioned-server-email-notifier",
	pubsub.SubscriptionConfig[*messages.ServerProvisioned]{
		Handler: func(ctx context.Context, msg *messages.ServerProvisioned) error {
			cfg := brevo.NewConfiguration()

			cfg.AddDefaultHeader("api-key", secrets.BrevoAPIKey)

			br := brevo.NewAPIClient(cfg)

			// mail
			// var params interface{}

			params := map[string]interface{}{
				"adminUrl":        "https://foo-bar.com",
				"serverIp":        "127.0.0.1",
				"hoursReserved":   "4",
				"terminationDate": "20.01.2025",
			}

			_, _, err := br.TransactionalEmailsApi.SendTransacEmail(ctx, brevo.SendSmtpEmail{
				To: []brevo.SendSmtpEmailTo{
					{
						Email: "anes.hasicic@gmail.com",
						Name:  "Customer",
					},
					{
						Email: "me@anes.io",
						Name:  "Admin",
					},
				},
				TemplateId: 1,
				Params:     params,
			})
			if err != nil {
				return err
			}

			_, _ = checkout.ServerDetails(ctx, msg.ServerID)
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

			_, _ = checkout.ServerDetails(ctx, msg.ServerID)

			// TODO - CC me - configurable

			return nil
		},
		MaxConcurrency: 20,
		RetryPolicy: &pubsub.RetryPolicy{
			MaxRetries: 20,
		},
	},
)
