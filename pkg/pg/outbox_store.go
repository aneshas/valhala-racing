package pg

import (
	"context"
	"database/sql"
	"encore.app/pkg/infra"
	"encore.app/pkg/messages"
	"encore.dev/pubsub"
	"fmt"
	"github.com/aneshas/tx/v2/sqltx"
	"x.encore.dev/infra/pubsub/outbox"
)

// NewOutboxStore creates a new outbox store
func NewOutboxStore(db *sql.DB, topicRefs infra.TopicRefs) *OutboxStore {
	return &OutboxStore{
		db:   db,
		refs: topicRefs,
	}
}

// OutboxStore represents an outbox postgres store for event publishing.
// Implements publisher
type OutboxStore struct {
	db   *sql.DB
	refs map[any]any
}

// Publish publishes events to the respective topics using encore outbox pattern
func (os *OutboxStore) Publish(ctx context.Context, events ...any) error {
	tx, ok := sqltx.From(ctx)
	if !ok {
		return fmt.Errorf("no tx present in context")
	}

	var err error

	pf := outbox.StdlibTxPersister(tx.Tx)

	for _, e := range events {
		// TODO - Generic?

		switch evt := e.(type) {
		case messages.ServerPaymentReceived:
			ref := outbox.Bind(topicRef[messages.ServerPaymentReceived](os.refs), pf)
			_, err = ref.Publish(ctx, &evt)

		case messages.ServerProvisioned:
			ref := outbox.Bind(topicRef[messages.ServerProvisioned](os.refs), pf)
			_, err = ref.Publish(ctx, &evt)

		case messages.ServerTerminationScheduled:
			ref := outbox.Bind(topicRef[messages.ServerTerminationScheduled](os.refs), pf)
			_, err = ref.Publish(ctx, &evt)

		case messages.ServerTerminated:
			ref := outbox.Bind(topicRef[messages.ServerTerminated](os.refs), pf)
			_, err = ref.Publish(ctx, &evt)
		}
	}

	return err
}

func topicRef[T any](m map[any]any) pubsub.Publisher[*T] {
	var t T

	return m[t].(pubsub.Publisher[*T])
}
