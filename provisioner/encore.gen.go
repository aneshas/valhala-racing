// Code generated by encore. DO NOT EDIT.

package provisioner

import "context"

// These functions are automatically generated and maintained by Encore
// to simplify calling them from other services, as they were implemented as methods.
// They are automatically updated by Encore whenever your API endpoints change.

// ScheduleTermination schedules a batch of expired servers for termination
func ScheduleTermination(ctx context.Context) error {
	// The implementation is elided here, and generated at compile-time by Encore.
	return nil
}

// Interface defines the service's API surface area, primarily for mocking purposes.
//
// Raw endpoints are currently excluded from this interface, as Encore does not yet
// support service-to-service API calls to raw endpoints.
type Interface interface {
	scheduleServerTermination(ctx context.Context) error

	// ScheduleTermination schedules a batch of expired servers for termination
	ScheduleTermination(ctx context.Context) error
}