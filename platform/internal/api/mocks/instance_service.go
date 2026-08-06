package mocks

import (
	"context"

	"github.com/foyez/dbaas-platform/platform/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockInstanceService struct {
	mock.Mock
}

func (m *MockInstanceService) CreateInstance(
	ctx context.Context,
	input domain.CreateInstanceInput,
) (*domain.CreateInstanceResult, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.CreateInstanceResult), args.Error(1)
}

func (m *MockInstanceService) ListInstances(ctx context.Context) (*domain.ListInstancesResult, error) {
	args := m.Called(ctx)

	if result := args.Get(0); result != nil {
		return result.(*domain.ListInstancesResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInstanceService) DeleteInstance(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
