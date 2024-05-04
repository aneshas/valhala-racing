package emailer

import (
	"context"
	"encore.app/checkout"
	"encore.app/pkg/messages"
	"encore.app/provisioner"
	brevo "github.com/getbrevo/brevo-go/lib"
)

// Service represents e-mailer api service
//
// encore:service
type Service struct {
	client *brevo.APIClient
}

// HandleServerProvisioned sends an email after a server has been provisioned
//
//encore:api private method=POST path=/emailer/handleServerProvisioned
func (s *Service) HandleServerProvisioned(ctx context.Context, msg *messages.ServerProvisioned) error {
	serverDetails, err := checkout.ServerDetails(ctx, msg.ServerID)
	if err != nil {
		return err
	}

	instanceDetails, err := provisioner.InstanceDetails(ctx, msg.InstanceID)
	if err != nil {
		return err
	}

	params := map[string]interface{}{
		"adminUrl":        instanceDetails.AdminURL,
		"serverIp":        instanceDetails.IPAddr,
		"hoursReserved":   serverDetails.HoursReserved,
		"terminationDate": instanceDetails.ExpiresOn,
	}

	_, _, err = s.client.TransactionalEmailsApi.SendTransacEmail(
		ctx,
		brevo.SendSmtpEmail{
			To: []brevo.SendSmtpEmailTo{
				{
					Email: serverDetails.UserEmail,
					Name:  "Customer",
				},
			},
			TemplateId: 1,
			Params:     params,
		})

	return err
}

// HandleServerTerminated sends an email after a server has been terminated
//
//encore:api private method=POST path=/emailer/handleServerTerminated
func (s *Service) HandleServerTerminated(ctx context.Context, msg *messages.ServerTerminated) error {
	serverDetails, err := checkout.ServerDetails(ctx, msg.ServerID)
	if err != nil {
		return err
	}

	_, _, err = s.client.TransactionalEmailsApi.SendTransacEmail(
		ctx,
		brevo.SendSmtpEmail{
			To: []brevo.SendSmtpEmailTo{
				{
					Email: serverDetails.UserEmail,
					Name:  "Customer",
				},
			},
			TemplateId: 2,
		})

	return err
}
