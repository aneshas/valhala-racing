package checkout

import (
	"context"
	"encore.app/checkout/pg"
	"encore.app/checkout/stripe"
	"encore.app/pkg/infra"
	"encore.app/pkg/messages"
	stdpg "encore.app/pkg/pg"
	encore "encore.dev"
	"encore.dev/pubsub"
	"encore.dev/storage/sqldb"
	"fmt"
	"github.com/aneshas/tx/v2"
	"github.com/aneshas/tx/v2/sqltx"
	"x.encore.dev/infra/pubsub/outbox"
)

var secrets struct {
	StripeApiKey         string
	StripeEndpointSecret string
}

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

	url := encore.Meta().APIBaseURL
	host := fmt.Sprintf("%s://%s", url.Scheme, url.Host)

	return &Service{
		Transactor: tx.New(sqltx.NewDB(stdDB)),
		payments:   stripe.NewGateway(secrets.StripeApiKey, secrets.StripeEndpointSecret, host),
		store:      pg.NewServerStore(stdDB),
		pub:        stdpg.NewOutboxStore(stdDB, refs),
	}, nil
}
