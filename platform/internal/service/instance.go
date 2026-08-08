// Package service contains application business logic implementations.
// Services orchestrate domain operations and coordinate dependencies
// such as repositories, Kubernetes clients, or external providers.
package service

import (
	"context"
	"log/slog"

	"github.com/foyez/dbaas-platform/platform/internal/domain"
)

// instanceService implements domain.InstanceService.
type instanceService struct {
	client domain.InstanceClient
	logger *slog.Logger
}

func NewInstanceService(
	client domain.InstanceClient,
	logger *slog.Logger,
) *instanceService {
	return &instanceService{
		client: client,
		logger: logger,
	}
}

// CreateInstance creates a new instance or returns an existing instance
// when the same idempotency key was already processed.
func (s *instanceService) CreateInstance(
	ctx context.Context,
	input domain.CreateInstanceInput,
) (*domain.CreateInstanceResult, error) {
	if err := validateCreateInstance(input); err != nil {
		return nil, err
	}

	instance, err := s.client.CreateInstance(ctx, input)
	if err != nil {
		return nil, err
	}

	return &domain.CreateInstanceResult{
		Instance: instance,
		Replayed: false,
	}, nil
}

func validateCreateInstance(
	input domain.CreateInstanceInput,
) error {
	if input.Name == "" {
		return ErrInvalidInput
	}
	if input.Version < 14 || input.Version > 16 {
		return ErrInvalidInput
	}
	if input.Storage == "" {
		return ErrInvalidInput
	}

	return nil
}

// GetInstance retrives a single instances by its ID.
func (s *instanceService) GetInstance(ctx context.Context, id string) (*domain.Instance, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}

	return s.client.GetInstance(ctx, id)
}

// ListInstances retrive list instances.
func (s *instanceService) ListInstances(ctx context.Context) (*domain.ListInstancesResult, error) {
	instances, err := s.client.ListInstances(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.ListInstancesResult{
		Instances: instances,
	}, nil
}

// UpdateInstance updates an instance.
func (s *instanceService) UpdateInstance(
	ctx context.Context,
	input domain.UpdateInstanceInput,
) (*domain.Instance, error) {
	if err := validateUpdateInstance(input); err != nil {
		return nil, err
	}

	return s.client.UpdateInstance(ctx, input)
}

func validateUpdateInstance(
	input domain.UpdateInstanceInput,
) error {
	if input.ID == "" {
		return ErrInvalidInput
	}
	if input.Version == nil && input.Storage == nil {
		return ErrNoUpdateFields
	}
	if input.Version != nil && (*input.Version < 14 || *input.Version > 16) {
		return ErrInvalidInput
	}
	if input.Storage != nil && *input.Storage == "" {
		return ErrInvalidInput
	}

	return nil
}

func (s *instanceService) DeleteInstance(ctx context.Context, id string) error {
	return s.client.DeleteInstance(ctx, id)
}
