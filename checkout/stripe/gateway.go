package stripe

import (
	"encoding/json"
	"encore.app/pkg/errs"
	"encore.dev/rlog"
	"errors"
	"fmt"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"
	"io"
	"net/http"
)

var prices = map[int]string{
	1:  "price_1ORCGxKZ3doyTEC4s3CyKSra",
	3:  "price_1ORCIWKZ3doyTEC4EcqVG9BS",
	5:  "price_1ORh37KZ3doyTEC4ZWmHSXUk",
	10: "price_1ORh4QKZ3doyTEC4MqLgYdPd",
}

func NewGateway(apiKey string, endpointSecret string, host string) *Gateway {
	stripe.Key = apiKey

	return &Gateway{
		APIKey:         apiKey,
		EndpointSecret: endpointSecret,
		Host:           host,
	}
}

type Gateway struct {
	APIKey         string
	EndpointSecret string
	Host           string
}

func (g *Gateway) StartSession(hoursRequested int, customerEmail string) (string, string, error) {
	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(prices[hoursRequested]),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:          stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:    stripe.String(g.Host + "/shop/success.html"),
		CancelURL:     stripe.String(g.Host + "/shop"),
		CustomerEmail: &customerEmail,
	}

	sess, err := session.New(params)
	if err != nil {
		rlog.Error(err.Error())

		return "", "", errors.Join(err, fmt.Errorf("could not create stripe session"))
	}

	return sess.ID, sess.URL, nil
}

func (g *Gateway) HandleCheckoutCompleted(w http.ResponseWriter, req *http.Request, h func(ref string) error) error {
	const maxBodyBytes = int64(65536)

	req.Body = http.MaxBytesReader(w, req.Body, maxBodyBytes)

	payload, err := io.ReadAll(req.Body)
	if err != nil {
		return errors.Join(err, errs.ErrTransientPaymentFailure)
	}

	// If you are testing your webhook locally with the Stripe CLI you
	// can find the endpoint's secret by running `stripe listen`
	// Otherwise, find your endpoint's secret in your webhook settings
	// in the Developer Dashboard

	// Pass the request body and Stripe-Signature header to ConstructEvent, along
	// with the webhook signing key.
	event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"), g.EndpointSecret)
	if err != nil {
		return err
	}

	switch event.Type {
	case "checkout.session.completed":
		var s stripe.CheckoutSession

		err := json.Unmarshal(event.Data.Raw, &s)
		if err != nil {
			return err
		}

		return h(s.ID)

	default:
		return nil
	}
}