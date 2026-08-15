package mocks

import (
	"context"

	"github.com/foyez/dbaas-platform/platform/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockInstanceClient is a mock implementation of domain.InstanceClient.
// It lets service-layer tests exercise business logic (validation,
// error propagation) without a real Kubernetes/CNPG dependency.
type MockInstanceClient struct {
	mock.Mock
}

func (m *MockInstanceClient) CreateInstance(
	ctx context.Context,
	input domain.CreateInstanceInput,
) (*domain.Instance, error) {
	args := m.Called(ctx, input)

	if result := args.Get(0); result != nil {
		return result.(*domain.Instance), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInstanceClient) GetInstanceByIdempotencyKey(
	ctx context.Context,
	key string,
) (*domain.Instance, error) {
	args := m.Called(ctx, key)

	if result := args.Get(0); result != nil {
		return result.(*domain.Instance), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInstanceClient) GetInstance(
	ctx context.Context,
	id string,
) (*domain.Instance, error) {
	args := m.Called(ctx, id)

	if result := args.Get(0); result != nil {
		return result.(*domain.Instance), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInstanceClient) ListInstances(
	ctx context.Context,
) ([]*domain.Instance, error) {
	args := m.Called(ctx)

	if result := args.Get(0); result != nil {
		return result.([]*domain.Instance), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInstanceClient) UpdateInstance(
	ctx context.Context,
	input domain.UpdateInstanceInput,
) (*domain.Instance, error) {
	args := m.Called(ctx, input)

	if result := args.Get(0); result != nil {
		return result.(*domain.Instance), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInstanceClient) DeleteInstance(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
