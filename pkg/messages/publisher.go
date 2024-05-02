package messages

import "context"

// Publisher represents a message publisher
type Publisher interface {
	Publish(ctx context.Context, events ...any) error
}
