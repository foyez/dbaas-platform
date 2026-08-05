// Package api contains HTTP transport layer types.
// DTOs (Data Transfer Objects) in this package define the request and response contracts
// exposed through the API. These types are used for JSON serialization
// and should not contain business logic.
package api

import (
	"time"

	"github.com/foyez/dbaas-platform/platform/internal/domain"
)

type CreateInstanceRequest struct {
	// Name      string `json:"name" binding:"required,min=3,max=63,dns1123"`
	Name    string `json:"name" binding:"required,min=3,max=63"`
	Version int    `json:"version" binding:"required,oneof=14 15 16"`
	// Storage   string `json:"storage" binding:"required,storage"`
	Storage   string `json:"storage" binding:"required"`
	Instances int    `json:"instances" binding:"omitempty,min=1,max=5"`
	Username  string `json:"username" binding:"omitempty,max=128"`
}

type InstanceResponse struct {
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	Version   int                   `json:"version"`
	Storage   string                `json:"storage"`
	Status    domain.InstanceStatus `json:"status"`
	CreatedAt time.Time             `json:"createdAt"`
}

type ListInstancesResponse struct {
	Items []InstanceResponse `json:"items"`
}
