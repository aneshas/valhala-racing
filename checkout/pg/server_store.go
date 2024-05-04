package pg

import (
	"context"
	"database/sql"
	"encore.app/checkout/pg/boiler"
	"encore.app/checkout/server"
	stdpg "encore.app/pkg/pg"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"time"
)

// NewServerStore instantiates a new server store
func NewServerStore(db *sql.DB) *ServerStore {
	return &ServerStore{db: db}
}

// ServerStore represents a postgres store for servers
type ServerStore struct {
	db *sql.DB
}

// Save saves a server
func (s *ServerStore) Save(ctx context.Context, server server.Server) (uint64, error) {
	conn := stdpg.Conn(ctx, s.db)

	entry := boiler.Server{
		UserEmail:     server.UserEmail,
		HoursReserved: server.HoursReserved,
		PaymentRef:    server.PaymentRef,
	}

	err := entry.Insert(ctx, conn, boil.Infer())
	if err != nil {
		return 0, err
	}

	return uint64(entry.ID), nil
}

// Update updates a server
func (s *ServerStore) Update(ctx context.Context, server server.Server) error {
	conn := stdpg.Conn(ctx, s.db)

	entry := boiler.Server{
		ID:                int64(server.ID),
		UserEmail:         server.UserEmail,
		HoursReserved:     server.HoursReserved,
		PaymentReceivedAt: null.TimeFromPtr(server.PaymentReceivedAt),
		PaymentRef:        server.PaymentRef,
		CreatedAt:         server.CreatedAt,
		UpdatedAt:         time.Now().UTC(),
	}

	_, err := entry.Update(ctx, conn, boil.Infer())

	return err
}

// FindByID finds server by id
func (s *ServerStore) FindByID(ctx context.Context, id uint64) (*server.Server, error) {
	conn := stdpg.Conn(ctx, s.db)

	result, err := boiler.FindServer(ctx, conn, int64(id))
	if err != nil {
		return nil, err
	}

	svr := fromBoilServer(result)

	return &svr, nil
}

// FindByPaymentRef finds a server by payment reference
func (s *ServerStore) FindByPaymentRef(ctx context.Context, paymentRef string) (*server.Server, error) {
	conn := stdpg.Conn(ctx, s.db)

	entry, err := boiler.Servers(
		boiler.ServerWhere.PaymentRef.EQ(paymentRef),
		boiler.ServerWhere.PaymentReceivedAt.IsNull(),
	).One(ctx, conn)
	if err != nil {
		return nil, err
	}

	svr := fromBoilServer(entry)

	return &svr, nil
}

func fromBoilServer(s *boiler.Server) server.Server {
	return server.Server{
		ID:                uint64(s.ID),
		UserEmail:         s.UserEmail,
		HoursReserved:     s.HoursReserved,
		PaymentRef:        s.PaymentRef,
		PaymentReceivedAt: s.PaymentReceivedAt.Ptr(),
		CreatedAt:         s.CreatedAt,
		UpdatedAt:         s.UpdatedAt,
	}
}
