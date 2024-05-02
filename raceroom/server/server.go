package server

import (
	"time"
)

// New instantiates a new server
func New(hoursReserved int, email string, paymentRef string) *Server {
	return &Server{
		HoursReserved: hoursReserved,
		UserEmail:     email,
		PaymentRef:    paymentRef,
	}
}

// Server represents RaceRoom dedicated server
type Server struct {
	ID                     uint64     `json:"id,omitempty"`
	UserEmail              string     `json:"userEmail,omitempty"`
	HoursReserved          int        `json:"hoursReserved,omitempty"`
	InstanceID             string     `json:"instanceId,omitempty"`
	PaymentRef             string     `json:"paymentRef,omitempty"`
	PaymentReceivedAt      *time.Time `json:"paymentReceivedAt,omitempty"`
	ProvisionedAt          *time.Time `json:"provisionedAt,omitempty"`
	TerminationScheduledAt *time.Time `json:"terminationScheduledAt,omitempty"`
	TerminatedAt           *time.Time `json:"terminatedAt,omitempty"`
	CreatedAt              time.Time  `json:"createdAt"`
	UpdatedAt              time.Time  `json:"updatedAt"`
}

// RegisterPayment marks the server as paid for
func (s *Server) RegisterPayment() {
	if s.PaymentReceivedAt != nil {
		return
	}

	now := time.Now().UTC()

	s.PaymentReceivedAt = &now
}
