package messages

// ServerPaymentReceived represents a server payment received event
type ServerPaymentReceived struct {
	ServerID      uint64 `json:"server_id"`
	HoursReserved int    `json:"hours_reserved"`
}

// ServerProvisioned represents a provisioned server event
type ServerProvisioned struct {
	ServerID   uint64 `json:"server_id"`
	InstanceID string `json:"instance_id"`
}

// ServerTerminationScheduled represents a scheduled server termination event
type ServerTerminationScheduled struct {
	ServerID   uint64 `json:"server_id"`
	InstanceID string `json:"instance_id"`
}

// ServerTerminated represents a terminated server event
type ServerTerminated struct {
	ServerID uint64 `json:"server_id"`
}
