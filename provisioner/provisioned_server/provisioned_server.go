package provisionedserver

import "time"

// New instantiates new provisioned server
func New(serverID uint64, instanceID string, hours int) *ProvisionedServer {
	expiry := time.Now().
		UTC().
		Add(time.Duration(hours) * time.Hour).
		Add(5 * time.Minute)

	return &ProvisionedServer{
		ServerID:   serverID,
		InstanceID: instanceID,
		ExpiresAt:  expiry,
	}
}

// ProvisionedServer represents a provisioned server
type ProvisionedServer struct {
	ID                     uint64     `json:"id,omitempty"`
	ServerID               uint64     `json:"serverId,omitempty"`
	InstanceID             string     `json:"instanceId,omitempty"`
	ExpiresAt              time.Time  `json:"expiresAt,omitempty"`
	TerminationScheduledAt *time.Time `json:"terminationScheduledAt,omitempty"`
	TerminatedAt           *time.Time `json:"terminatedAt,omitempty"`
	CreatedAt              time.Time  `json:"createdAt"`
	UpdatedAt              time.Time  `json:"updatedAt"`
}

// ScheduleTermination marks server as scheduled for termination
func (s *ProvisionedServer) ScheduleTermination() {
	now := time.Now().UTC()

	s.TerminationScheduledAt = &now
}
