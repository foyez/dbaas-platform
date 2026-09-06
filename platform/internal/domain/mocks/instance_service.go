package mocks

import (
	"context"

	"github.com/foyez/dbaas-platform/platform/internal/domain"
	"github.com/foyez/dbaas-platform/platform/internal/infra/loki"
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

func (m *MockInstanceService) GetInstance(ctx context.Context, id string) (*domain.Instance, error) {
	args := m.Called(ctx, id)

	if result := args.Get(0); result != nil {
		return result.(*domain.Instance), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInstanceService) ListInstances(
	ctx context.Context,
	userID string,
	limit int,
	cursor string,
) (*domain.ListInstancesResult, error) {
	args := m.Called(ctx, userID, limit, cursor)

	if result := args.Get(0); result != nil {
		return result.(*domain.ListInstancesResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInstanceService) UpdateInstance(
	ctx context.Context,
	input domain.UpdateInstanceInput,
) (*domain.Instance, error) {
	args := m.Called(ctx, input)

	if result := args.Get(0); result != nil {
		return result.(*domain.Instance), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInstanceService) DeleteInstance(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockInstanceService) GetAuditLogs(
	ctx context.Context,
	resourceID string,
	limit int,
	cursor string,
) ([]loki.LogLine, string, error) {
	args := m.Called(ctx, resourceID, limit, cursor)

	var logs []loki.LogLine
	if result := args.Get(0); result != nil {
		logs = result.([]loki.LogLine)
	}

	return logs, args.String(1), args.Error(2)
}
