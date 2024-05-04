package checkout

import (
	"context"
	"encore.app/checkout/server"
	"encore.app/pkg/errs"
	"encore.app/pkg/messages"
	"encore.dev/rlog"
	"errors"
	"github.com/aneshas/tx/v2"
	"net/http"
)

// PaymentProvider represents external payment gateway
type PaymentProvider interface {
	StartSession(hoursRequested int, customerEmail string) (string, string, error)
	HandleCheckoutCompleted(w http.ResponseWriter, req *http.Request, h func(ref string) error) error
}

// Store represents a store for servers
type Store interface {
	Save(ctx context.Context, server server.Server) (uint64, error)
	Update(ctx context.Context, server server.Server) error
	FindByID(ctx context.Context, id uint64) (*server.Server, error)
	FindByPaymentRef(ctx context.Context, paymentRef string) (*server.Server, error)
}

// Service represents checkout api service
//
// encore:service
type Service struct {
	tx.Transactor

	payments PaymentProvider
	store    Store
	pub      messages.Publisher
}

// RequestServerReq represents a requested server api request
type RequestServerReq struct {
	HoursReserved int    `json:"hoursReserved"`
	Email         string `json:"email"`
}

// RequestServerResp represents a requested server api response
type RequestServerResp struct {
	ServerID        uint64 `json:"serverId"`
	PaymentRedirect string `json:"paymentRedirect"`
}

// RequestServer requests a new server
//
// encore:api private method=POST path=/checkout/server/request
func (s *Service) RequestServer(ctx context.Context, req *RequestServerReq) (resp *RequestServerResp, err error) {
	return resp, s.WithTransaction(ctx, func(ctx context.Context) error {
		ref, checkoutURL, err := s.payments.StartSession(req.HoursReserved, req.Email)
		if err != nil {
			return err
		}

		svr := server.New(req.HoursReserved, req.Email, ref)

		id, err := s.store.Save(ctx, *svr)
		if err != nil {
			return err
		}

		resp = &RequestServerResp{
			ServerID:        id,
			PaymentRedirect: checkoutURL,
		}

		return nil
	})
}

//encore:api public raw path=/checkout/payment-callback
func (s *Service) PaymentCallback(w http.ResponseWriter, req *http.Request) {
	err := s.payments.HandleCheckoutCompleted(w, req, func(ref string) error {
		return RegisterServerPayment(req.Context(), &RegisterServerPaymentReq{
			PaymentRef: ref,
		})
	})
	if err != nil {
		rlog.Error(err.Error())

		if errors.Is(err, errs.ErrTransientPaymentFailure) {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RegisterServerPaymentReq represents register server payment request
type RegisterServerPaymentReq struct {
	PaymentRef string
}

// PaymentReceived represents register server payment request
type PaymentReceived struct {
	ServerID uint64 `json:"server_id"`
}

// RegisterServerPayment marks server as paid
//
//encore:api private method=POST path=/checkout/server/registerPayment
func (s *Service) RegisterServerPayment(ctx context.Context, req *RegisterServerPaymentReq) error {
	return s.WithTransaction(ctx, func(ctx context.Context) error {
		svr, err := s.store.FindByPaymentRef(ctx, req.PaymentRef)
		if err != nil {
			return err
		}

		svr.RegisterPayment()

		err = s.store.Update(ctx, *svr)
		if err != nil {
			return err
		}

		return s.pub.Publish(ctx, messages.ServerPaymentReceived{
			ServerID:      svr.ID,
			HoursReserved: svr.HoursReserved,
		})
	})
}

//encore:api private method=GET path=/checkout/server/details/:id
func (s *Service) ServerDetails(ctx context.Context, id uint64) (*server.Server, error) {
	return s.store.FindByID(ctx, id)
}
