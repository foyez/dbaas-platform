package mocks

import (
	"context"

	"github.com/foyez/dbaas-platform/platform/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockQuerier is a mock implementation of ports.Querier.
type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) ExecuteQuery(
	ctx context.Context,
	instanceID string,
	creds *domain.InstanceCredentials,
	sql string,
) (*domain.QueryResult, error) {
	args := m.Called(ctx, instanceID, creds, sql)

	if result := args.Get(0); result != nil {
		return result.(*domain.QueryResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockQuerier) Close() {
	m.Called()
}
