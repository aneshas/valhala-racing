package provisioner

import (
	"context"
	"encore.app/pkg/infra"
	"encore.app/pkg/messages"
	stdpg "encore.app/pkg/pg"
	"encore.app/provisioner/pg"
	"encore.dev/cron"
	"encore.dev/pubsub"
	"encore.dev/storage/sqldb"
	"github.com/aneshas/tx/v2"
	"github.com/aneshas/tx/v2/sqltx"
	"x.encore.dev/infra/pubsub/outbox"
)

var db = sqldb.NewDatabase(
	"provisioner",
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
		ref := pubsub.TopicRef[pubsub.Publisher[*messages.ServerProvisioned]](infra.ProvisionedServersTopic)
		outbox.RegisterTopic(relay, ref)
		refs[messages.ServerProvisioned{}] = ref
	}

	{
		ref := pubsub.TopicRef[pubsub.Publisher[*messages.ServerTerminationScheduled]](infra.ScheduledServerTerminationsTopic)
		outbox.RegisterTopic(relay, ref)
		refs[messages.ServerTerminationScheduled{}] = ref
	}

	{
		ref := pubsub.TopicRef[pubsub.Publisher[*messages.ServerTerminated]](infra.TerminatedServersTopic)
		outbox.RegisterTopic(relay, ref)
		refs[messages.ServerTerminated{}] = ref
	}

	go relay.PollForMessages(context.Background(), -1)

	return &Service{
		Transactor: tx.New(sqltx.NewDB(stdDB)),
		store:      pg.NewProvisionedServerStore(stdDB),
		pub:        stdpg.NewOutboxStore(stdDB, refs),
	}, nil
}

var _ = pubsub.NewSubscription(
	infra.ReceivedServerPaymentsTopic,
	"aws-server-provisioner",
	pubsub.SubscriptionConfig[*messages.ServerPaymentReceived]{
		Handler:        pubsub.MethodHandler((*Service).HandlePaymentReceived),
		MaxConcurrency: 20,
		RetryPolicy: &pubsub.RetryPolicy{
			MaxRetries: 20,
		},
	},
)

var _ = pubsub.NewSubscription(
	infra.ScheduledServerTerminationsTopic,
	"aws-server-terminator",
	pubsub.SubscriptionConfig[*messages.ServerTerminationScheduled]{
		Handler:        pubsub.MethodHandler((*Service).HandleScheduledTermination),
		MaxConcurrency: 20,
		RetryPolicy: &pubsub.RetryPolicy{
			MaxRetries: 20,
		},
	},
)

var _ = cron.NewJob("provisioner-termination-scheduler", cron.JobConfig{
	Title:    "Schedule expired server termination",
	Every:    cron.Minute,
	Endpoint: scheduleServerTermination,
})

//encore:api private
func scheduleServerTermination(ctx context.Context) error {
	return ScheduleTermination(ctx)
}