package raceroom

import (
	"context"
	"encore.app/pkg/messages"
	"encore.app/raceroom/server"
	"github.com/aneshas/tx/v2"
)

// PaymentProvider represents external payment gateway
type PaymentProvider interface {
	StartSession(hoursRequested int, customerEmail string) (string, string, error)
}

// Store represents a store for servers
type Store interface {
	Save(ctx context.Context, server server.Server) (uint64, error)
	Update(ctx context.Context, server server.Server) error
	FindByPaymentRef(ctx context.Context, paymentRef string) (*server.Server, error)
}

// Service represents RaceRoom api service
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
// encore:api private method=POST path=/raceRoom/server/request
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

type RegisterServerPaymentReq struct {
	PaymentRef string
}

type PaymentReceived struct {
	ServerID uint64 `json:"server_id"`
}

// RegisterServerPayment marks server as paid
//
//encore:api private method=POST path=/raceRoom/server/registerPayment
func (s *Service) RegisterServerPayment(ctx context.Context, req *RegisterServerPaymentReq) error {
	return s.WithTransaction(ctx, func(ctx context.Context) error {
		svr, err := s.store.FindByPaymentRef(ctx, req.PaymentRef)
		if err != nil {
			return err
		}

		svr.RegisterPayment()

		err = s.pub.Publish(ctx, messages.ServerPaymentReceived{
			ServerID:      svr.ID,
			HoursReserved: svr.HoursReserved,
		})
		if err != nil {
			return err
		}

		return s.store.Update(ctx, *svr)
	})
}

//encore:api private method=GET path=/raceRoom/server/details/:id
func (s *Service) ServerDetails(ctx context.Context, id uint64) (*server.Server, error) {
	return &server.Server{
		UserEmail: "mail@mail.com",
	}, nil
}
