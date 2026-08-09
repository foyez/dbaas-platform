package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/foyez/dbaas-platform/platform/internal/domain"
	"github.com/foyez/dbaas-platform/platform/internal/domain/mocks"
	"github.com/foyez/dbaas-platform/platform/internal/logger"
	"github.com/foyez/dbaas-platform/platform/internal/service"
	"github.com/go-openapi/testify/v2/require"
	"github.com/stretchr/testify/mock"
)

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }

func newTestService(client *mocks.MockInstanceClient) domain.InstanceService {
	log := logger.New("debug", "test")
	return service.NewInstanceService(client, log)
}

func TestInstanceService_CreateIsntance(t *testing.T) {
	validInput := domain.CreateInstanceInput{
		Name:           "demo",
		Version:        16,
		Storage:        "10Gi",
		Username:       "postgres",
		IdempotencyKey: "key-1",
	}

	tests := []struct {
		name      string
		input     domain.CreateInstanceInput
		mockSetup func(client *mocks.MockInstanceClient)
		wantErr   error
		wantCall  bool
	}{
		{
			name:  "success",
			input: validInput,
			mockSetup: func(client *mocks.MockInstanceClient) {
				client.On("CreateInstance", mock.Anything, validInput).
					Return(&domain.Instance{ID: "123", Name: "domain"}, nil)
			},
			wantCall: true,
		},
		{
			name:      "missing name",
			input:     domain.CreateInstanceInput{Version: 16, Storage: "10Gi"},
			mockSetup: func(client *mocks.MockInstanceClient) {},
			wantErr:   service.ErrInvalidInput,
		},
		{
			name:      "version out of range",
			input:     domain.CreateInstanceInput{Name: "demo", Version: 99, Storage: "10Gi"},
			mockSetup: func(client *mocks.MockInstanceClient) {},
			wantErr:   service.ErrInvalidInput,
		},
		{
			name:      "missing storage",
			input:     domain.CreateInstanceInput{Name: "demo", Version: 16},
			mockSetup: func(client *mocks.MockInstanceClient) {},
			wantErr:   service.ErrInvalidInput,
		},
		{
			name:  "client error propagates",
			input: validInput,
			mockSetup: func(client *mocks.MockInstanceClient) {
				client.On("CreateInstance", mock.Anything, validInput).
					Return(nil, service.ErrAlreadyExists)
			},
			wantErr:  service.ErrAlreadyExists,
			wantCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := new(mocks.MockInstanceClient)
			tt.mockSetup(client)

			svc := newTestService(client)
			result, err := svc.CreateInstance(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}

			if tt.wantCall {
				client.AssertExpectations(t)
			} else {
				client.AssertNotCalled(t, "CreateInstance", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestInstanceService_GetInstance(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		mockSetup func(client *mocks.MockInstanceClient)
		wantErr   error
		wantCall  bool
	}{
		{
			name: "success",
			id:   "123",
			mockSetup: func(client *mocks.MockInstanceClient) {
				client.On("GetInstance", mock.Anything, "123").
					Return(&domain.Instance{ID: "123", Name: "demo"}, nil)
			},
			wantCall: true,
		},
		{
			name:      "empty id",
			id:        "",
			mockSetup: func(client *mocks.MockInstanceClient) {},
			wantErr:   service.ErrInvalidInput,
		},
		{
			name: "not found propagates",
			id:   "missing",
			mockSetup: func(client *mocks.MockInstanceClient) {
				client.On("GetInstance", mock.Anything, "missing").
					Return(nil, service.ErrNotFound)
			},
			wantErr:  service.ErrNotFound,
			wantCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := new(mocks.MockInstanceClient)
			tt.mockSetup(client)

			svc := newTestService(client)
			instance, err := svc.GetInstance(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, instance)
			} else {
				require.NoError(t, err)
				require.NotNil(t, instance)
			}

			if tt.wantCall {
				client.AssertExpectations(t)
			} else {
				client.AssertNotCalled(t, "GetInstance", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestInstanceService_ListInstances(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := new(mocks.MockInstanceClient)
		client.On("ListInstances", mock.Anything).
			Return([]*domain.Instance{{ID: "1"}, {ID: "2"}}, nil)

		svc := newTestService(client)
		result, err := svc.ListInstances(context.Background())

		require.NoError(t, err)
		require.Len(t, result.Instances, 2)
		client.AssertExpectations(t)
	})

	t.Run("client error propagates", func(t *testing.T) {
		client := new(mocks.MockInstanceClient)
		client.On("ListInstances", mock.Anything).
			Return(nil, errors.New("boom"))

		svc := newTestService(client)
		result, err := svc.ListInstances(context.Background())

		require.Error(t, err)
		require.Nil(t, result)
		client.AssertExpectations(t)
	})
}

func TestInstanceService_UpdateInstance(t *testing.T) {
	tests := []struct {
		name      string
		input     domain.UpdateInstanceInput
		mockSetup func(client *mocks.MockInstanceClient)
		wantErr   error
		wantCall  bool
	}{
		{
			name:  "success storage resize",
			input: domain.UpdateInstanceInput{ID: "123", Storage: strPtr("20Gi")},
			mockSetup: func(client *mocks.MockInstanceClient) {
				client.On("UpdateInstance", mock.Anything, mock.Anything).
					Return(&domain.Instance{ID: "123", Storage: "20Gi"}, nil)
			},
			wantCall: true,
		},
		{
			name:  "success version upgrade",
			input: domain.UpdateInstanceInput{ID: "123", Version: intPtr(16)},
			mockSetup: func(client *mocks.MockInstanceClient) {
				client.On("UpdateInstance", mock.Anything, mock.Anything).
					Return(&domain.Instance{ID: "123", Version: 16}, nil)
			},
			wantCall: true,
		},
		{
			name:      "empty id",
			input:     domain.UpdateInstanceInput{Storage: strPtr("20Gi")},
			mockSetup: func(client *mocks.MockInstanceClient) {},
			wantErr:   service.ErrInvalidInput,
		},
		{
			name:      "no fields provided",
			input:     domain.UpdateInstanceInput{ID: "123"},
			mockSetup: func(client *mocks.MockInstanceClient) {},
			wantErr:   service.ErrNoUpdateFields,
		},
		{
			name:      "version out of range",
			input:     domain.UpdateInstanceInput{ID: "123", Version: intPtr(99)},
			mockSetup: func(client *mocks.MockInstanceClient) {},
			wantErr:   service.ErrInvalidInput,
		},
		{
			name:      "empty storage string",
			input:     domain.UpdateInstanceInput{ID: "123", Storage: strPtr("")},
			mockSetup: func(client *mocks.MockInstanceClient) {},
			wantErr:   service.ErrInvalidInput,
		},
		{
			name:  "not found propagates",
			input: domain.UpdateInstanceInput{ID: "missing", Storage: strPtr("20Gi")},
			mockSetup: func(client *mocks.MockInstanceClient) {
				client.On("UpdateInstance", mock.Anything, mock.Anything).
					Return(nil, service.ErrNotFound)
			},
			wantErr:  service.ErrNotFound,
			wantCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := new(mocks.MockInstanceClient)
			tt.mockSetup(client)

			svc := newTestService(client)
			instance, err := svc.UpdateInstance(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, instance)
			} else {
				require.NoError(t, err)
				require.NotNil(t, instance)
			}

			if tt.wantCall {
				client.AssertExpectations(t)
			} else {
				client.AssertNotCalled(t, "UpdateInstance", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestInstanceService_DeleteInstance(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		mockSetup func(client *mocks.MockInstanceClient)
		wantErr   error
		wantCall  bool
	}{
		{
			name: "success",
			id:   "123",
			mockSetup: func(client *mocks.MockInstanceClient) {
				client.On("DeleteInstance", mock.Anything, "123").Return(nil)
			},
			wantCall: true,
		},
		{
			name:      "empty id",
			id:        "",
			mockSetup: func(client *mocks.MockInstanceClient) {},
			wantErr:   service.ErrInvalidInput,
		},
		{
			name: "not found propagates",
			id:   "missing",
			mockSetup: func(client *mocks.MockInstanceClient) {
				client.On("DeleteInstance", mock.Anything, "missing").
					Return(service.ErrNotFound)
			},
			wantErr:  service.ErrNotFound,
			wantCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := new(mocks.MockInstanceClient)
			tt.mockSetup(client)

			svc := newTestService(client)
			err := svc.DeleteInstance(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tt.wantCall {
				client.AssertExpectations(t)
			} else {
				client.AssertNotCalled(t, "DeleteInstance", mock.Anything, mock.Anything)
			}
		})
	}
}
