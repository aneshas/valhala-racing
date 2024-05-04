package pg

import (
	"context"
	"database/sql"
	stdpg "encore.app/pkg/pg"
	"encore.app/provisioner/pg/boiler"
	provisionedserver "encore.app/provisioner/provisioned_server"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"time"
)

// NewProvisionedServerStore instantiates a new server store
func NewProvisionedServerStore(db *sql.DB) *ProvisionedServerStore {
	return &ProvisionedServerStore{db: db}
}

// ProvisionedServerStore represents a postgres store for provisioned servers
type ProvisionedServerStore struct {
	db *sql.DB
}

// Save saves a provisioned server
func (s *ProvisionedServerStore) Save(ctx context.Context, server provisionedserver.ProvisionedServer) (uint64, error) {
	conn := stdpg.Conn(ctx, s.db)

	entry := boiler.ProvisionedServer{
		ServerID:   int(server.ServerID),
		InstanceID: server.InstanceID,
		ExpiresAt:  server.ExpiresAt,
	}

	err := entry.Insert(ctx, conn, boil.Infer())
	if err != nil {
		return 0, err
	}

	return uint64(entry.ID), nil
}

// Update updates a provisioned server
func (s *ProvisionedServerStore) Update(ctx context.Context, server provisionedserver.ProvisionedServer) error {
	conn := stdpg.Conn(ctx, s.db)

	entry := boiler.ProvisionedServer{
		ID:                     int64(server.ID),
		ServerID:               int(server.ServerID),
		ExpiresAt:              server.ExpiresAt,
		TerminationScheduledAt: null.TimeFromPtr(server.TerminationScheduledAt),
		TerminatedAt:           null.TimeFromPtr(server.TerminatedAt),
		CreatedAt:              server.CreatedAt,
		UpdatedAt:              time.Now().UTC(),
		InstanceID:             server.InstanceID,
	}

	_, err := entry.Update(ctx, conn, boil.Infer())

	return err
}

// FindByID finds server by id
func (s *ProvisionedServerStore) FindByID(ctx context.Context, id uint64) (*provisionedserver.ProvisionedServer, error) {
	conn := stdpg.Conn(ctx, s.db)

	result, err := boiler.FindProvisionedServer(ctx, conn, int64(id))
	if err != nil {
		return nil, err
	}

	svr := fromBoilServer(result)

	return &svr, nil
}

// select
// id,
// instance_id
// from server
// where
// server_provisioned_on is not null
// and scheduled_termination_on is null
// and server_terminated_on is null
// and AGE((now() at time zone 'utc'), server_provisioned_on) >
// (INTERVAL '1 hour' * hours_reserved + '5 minutes')
// for update nowait
// limit $1`,
// 		maxConcurrency,

// FindExpired returns a batch of expired provisioned servers
func (s *ProvisionedServerStore) FindExpired(ctx context.Context, limit int) ([]provisionedserver.ProvisionedServer, error) {
	conn := stdpg.Conn(ctx, s.db)

	result, err := boiler.ProvisionedServers(
		boiler.ProvisionedServerWhere.TerminationScheduledAt.IsNull(),
		boiler.ProvisionedServerWhere.TerminatedAt.IsNull(),
		// TODO - Other queries - eg. expired

	).All(ctx, conn)
	if err != nil {
		return nil, err
	}

	var servers []provisionedserver.ProvisionedServer

	for _, bs := range result {
		servers = append(servers, fromBoilServer(bs))
	}

	return servers, nil
}

func fromBoilServer(s *boiler.ProvisionedServer) provisionedserver.ProvisionedServer {
	return provisionedserver.ProvisionedServer{
		ID:                     uint64(s.ID),
		ServerID:               uint64(s.ServerID),
		InstanceID:             s.InstanceID,
		ExpiresAt:              s.ExpiresAt,
		TerminationScheduledAt: s.TerminationScheduledAt.Ptr(),
		TerminatedAt:           s.TerminatedAt.Ptr(),
		CreatedAt:              s.CreatedAt,
		UpdatedAt:              s.UpdatedAt,
	}
}
