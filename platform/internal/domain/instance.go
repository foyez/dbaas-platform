// Package domain contains core business models and rules.
// Types in this package are independent of external concerns such as
// HTTP, JSON, databases, or infrastructure providers.
package domain

import "time"

type InstanceStatus string

const (
	InstanceStatusPending InstanceStatus = "Pending"
	InstanceStatusRunning InstanceStatus = "Running"
	InstanceStatusFailed  InstanceStatus = "Failed"
	InstanceStatusDeleted InstanceStatus = "Deleted"
)

// Instance represents a provisioned PostgreSQL instance
// and its current lifecycle state.
type Instance struct {
	ID        string
	Name      string
	Version   int
	Storage   string
	Status    InstanceStatus
	CreatedAt time.Time
}

type CreateInstanceInput struct {
	Name           string
	Version        int
	Storage        string
	Username       string
	IdempotencyKey string
	UserID         string
}

type CreateInstanceResult struct {
	Instance *Instance
	Replayed bool
}

type ListInstancesResult struct {
	Instances []*Instance
}

type UpdateInstanceInput struct {
	ID      string
	Version *int
	Storage *string
}
