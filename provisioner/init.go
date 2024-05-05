package provisioner

import (
	"context"
	"encore.app/pkg/infra"
	"encore.app/pkg/messages"
	stdpg "encore.app/pkg/pg"
	"encore.app/provisioner/awsprovisioner"
	"encore.app/provisioner/pg"
	"encore.dev/config"
	"encore.dev/cron"
	"encore.dev/pubsub"
	"encore.dev/storage/cache"
	"encore.dev/storage/sqldb"
	"github.com/aneshas/tx/v2"
	"github.com/aneshas/tx/v2/sqltx"
	"github.com/friendsofgo/errors"
	"time"
	"x.encore.dev/infra/pubsub/outbox"
)

var secrets struct {
	AWSKeyID     string
	AWSKeySecret string
	AWSRoleARN   string
}

var db = sqldb.NewDatabase(
	"provisioner",
	sqldb.DatabaseConfig{
		Migrations: "./pg/migrations",
	},
)

type provisionerConfig struct {
	IsTest config.Bool
}

var cfg = config.Load[*provisionerConfig]()

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

	var provisioner ServerProvisioner = awsprovisioner.TestProvisioner

	if !cfg.IsTest() {
		provisioner = awsprovisioner.NewAWSProvisioner(secrets.AWSKeyID, secrets.AWSKeySecret, secrets.AWSRoleARN)
	}

	return &Service{
		Transactor:       tx.New(sqltx.NewDB(stdDB)),
		provisioner:      provisioner,
		isMessageSeenFor: cacheKeyExists,
		store:            pg.NewProvisionedServerStore(stdDB),
		pub:              stdpg.NewOutboxStore(stdDB, refs),
	}, nil
}

var _ = pubsub.NewSubscription(
	infra.ReceivedServerPaymentsTopic,
	"server-provisioner",
	pubsub.SubscriptionConfig[*messages.ServerPaymentReceived]{
		Handler:        pubsub.MethodHandler((*Service).HandlePaymentReceived),
		MaxConcurrency: 10,
		RetryPolicy: &pubsub.RetryPolicy{
			MaxRetries: 10,
		},
	},
)

var _ = pubsub.NewSubscription(
	infra.ScheduledServerTerminationsTopic,
	"server-terminator",
	pubsub.SubscriptionConfig[*messages.ServerTerminationScheduled]{
		Handler:        pubsub.MethodHandler((*Service).HandleScheduledTermination),
		MaxConcurrency: 10,
		RetryPolicy: &pubsub.RetryPolicy{
			MaxRetries: 10,
		},
	},
)

var _ = cron.NewJob("provisioner-termination-scheduler", cron.JobConfig{
	Title:    "Schedule expired server termination",
	Every:    cron.Minute * 5,
	Endpoint: scheduleServerTermination,
})

//encore:api private
func scheduleServerTermination(ctx context.Context) error {
	return ScheduleTermination(ctx)
}

var cacheCluster = cache.NewCluster("provisioner-cache-cluster", cache.ClusterConfig{
	EvictionPolicy: cache.AllKeysLRU,
})

var processedMessagesCache = cache.NewIntKeyspace[string](cacheCluster, cache.KeyspaceConfig{
	KeyPattern:    "provisioner-messages/:key",
	DefaultExpiry: cache.ExpireIn(10 * time.Minute),
})

func cacheKeyExists(ctx context.Context, key string) (bool, error) {
	val, err := processedMessagesCache.GetAndSet(ctx, key, 1)
	if err != nil {
		if errors.Is(err, cache.Miss) {
			return false, nil
		}

		return false, err
	}

	return val != 0, nil
}
