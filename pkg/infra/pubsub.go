package infra

import (
	"encore.app/pkg/messages"
	"encore.dev/pubsub"
)

// ReceivedServerPaymentsTopic is a topic for received server payments
var ReceivedServerPaymentsTopic = pubsub.NewTopic[*messages.ServerPaymentReceived](
	"received-server-payments-topic",
	pubsub.TopicConfig{
		DeliveryGuarantee: pubsub.ExactlyOnce,
	},
)

// ProvisionedServersTopic is a topic for provisioned servers
var ProvisionedServersTopic = pubsub.NewTopic[*messages.ServerProvisioned](
	"provisioned-servers-topic",
	pubsub.TopicConfig{
		DeliveryGuarantee: pubsub.AtLeastOnce,
	},
)

// ScheduledServerTerminationsTopic is a topic for scheduled server terminations
var ScheduledServerTerminationsTopic = pubsub.NewTopic[*messages.ServerTerminationScheduled](
	"scheduled-server-terminations-topic",
	pubsub.TopicConfig{
		DeliveryGuarantee: pubsub.AtLeastOnce,
	},
)

// TerminatedServersTopic is a topic for terminated servers
var TerminatedServersTopic = pubsub.NewTopic[*messages.ServerTerminated](
	"terminated-servers-topic",
	pubsub.TopicConfig{
		DeliveryGuarantee: pubsub.AtLeastOnce,
	},
)

// TopicRefs represents a topic ref lookup map
type TopicRefs map[any]any
