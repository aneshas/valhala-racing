package stripe

import (
	"encore.dev/rlog"
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

var prices = map[int]string{
	1:  "price_1ORCGxKZ3doyTEC4s3CyKSra",
	3:  "price_1ORCIWKZ3doyTEC4EcqVG9BS",
	5:  "price_1ORh37KZ3doyTEC4ZWmHSXUk",
	10: "price_1ORh4QKZ3doyTEC4MqLgYdPd",
}

func NewGateway(apiKey string) *Gateway {
	stripe.Key = apiKey

	return &Gateway{}
}

type Gateway struct {
}

func (g *Gateway) StartSession(hoursRequested int, customerEmail string) (string, string, error) {
	// TODO - Config
	domain := "http://localhost:4000"

	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(prices[hoursRequested]),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:          stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:    stripe.String(domain + "/shop/success.html"),
		CancelURL:     stripe.String(domain + "/shop"),
		CustomerEmail: &customerEmail,
	}

	sess, err := session.New(params)
	if err != nil {
		rlog.Error(err.Error())

		return "", "", errors.Join(err, fmt.Errorf("could not create stripe session"))
	}

	return sess.ID, sess.URL, nil
}
