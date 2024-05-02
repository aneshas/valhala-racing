package checkout

import (
	"context"
	"encore.app/checkout/pg"
	"encore.app/checkout/stripe"
	"encore.app/pkg/infra"
	"encore.app/pkg/messages"
	stdpg "encore.app/pkg/pg"
	"encore.dev/config"
	"encore.dev/pubsub"
	"encore.dev/storage/sqldb"
	"github.com/aneshas/tx/v2"
	"github.com/aneshas/tx/v2/sqltx"
	"x.encore.dev/infra/pubsub/outbox"
)

var secrets struct {
	StripeApiKey         string
	StripeEndpointSecret string
}

type checkoutConfig struct {
	Host config.String
}

var cfg = config.Load[*checkoutConfig]()

var db = sqldb.NewDatabase(
	"checkout",
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
		payments:   stripe.NewGateway(secrets.StripeApiKey, secrets.StripeEndpointSecret, cfg.Host()),
		store:      pg.NewServerStore(stdDB),
		pub:        stdpg.NewOutboxStore(stdDB, refs),
	}, nil
}
