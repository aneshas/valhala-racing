package raceroom

import (
	"context"
	"encore.app/pkg/infra"
	"encore.app/pkg/messages"
	stdpg "encore.app/pkg/pg"
	"encore.app/raceroom/pg"
	"encore.app/raceroom/stripe"
	"encore.dev/pubsub"
	"encore.dev/storage/sqldb"
	"github.com/aneshas/tx/v2"
	"github.com/aneshas/tx/v2/sqltx"
	"x.encore.dev/infra/pubsub/outbox"
)

var db = sqldb.NewDatabase(
	"raceroom",
	sqldb.DatabaseConfig{
		Migrations: "./pg/migrations",
	},
)

func initService() (*Service, error) {
	var (
		stdDB = db.Stdlib()
		relay = outbox.NewRelay(outbox.SQLDBStore(db))
		refs  = make(infra.TopicRefs)
	)

	{
		ref := pubsub.TopicRef[pubsub.Publisher[*messages.ServerPaymentReceived]](infra.ReceivedServerPaymentsTopic)
		outbox.RegisterTopic(relay, ref)
		refs[messages.ServerPaymentReceived{}] = ref
	}

	go relay.PollForMessages(context.Background(), -1)

	return &Service{
		Transactor: tx.New(sqltx.NewDB(stdDB)),
		payments:   stripe.NewGateway("sk_test_51ORBzOKZ3doyTEC4HiZhssbvxNYunSY08TdEncGuf2ZA2yqNmlmaXetOUogAXyOWq3osfUiHIXGATT2tqcbxkmNL00wb6olEQn"),
		store:      pg.NewServerStore(stdDB),
		pub:        stdpg.NewOutboxStore(stdDB, refs),
	}, nil
}
