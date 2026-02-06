// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util/passhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStore implements storage.Storage for testing
type MockStore struct {
	mock.Mock
}

func (m *MockStore) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.McpAnyServerConfig), args.Error(1)
}

func (m *MockStore) Watch(ctx context.Context) (<-chan *configv1.McpAnyServerConfig, error) {
	args := m.Called(ctx)
	return nil, args.Error(1)
}

func (m *MockStore) ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.UpstreamServiceConfig), args.Error(1)
}

func (m *MockStore) GetService(ctx context.Context, id string) (*configv1.UpstreamServiceConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.UpstreamServiceConfig), args.Error(1)
}

func (m *MockStore) SaveService(ctx context.Context, svc *configv1.UpstreamServiceConfig) error {
	args := m.Called(ctx, svc)
	return args.Error(0)
}

func (m *MockStore) DeleteService(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) GetGlobalSettings(ctx context.Context) (*configv1.GlobalSettings, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.GlobalSettings), args.Error(1)
}

// Secrets
func (m *MockStore) ListSecrets(ctx context.Context) ([]*configv1.Secret, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.Secret), args.Error(1)
}

func (m *MockStore) GetSecret(ctx context.Context, id string) (*configv1.Secret, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.Secret), args.Error(1)
}

func (m *MockStore) SaveSecret(ctx context.Context, secret *configv1.Secret) error {
	args := m.Called(ctx, secret)
	return args.Error(0)
}

func (m *MockStore) DeleteSecret(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Dashboard Layouts
func (m *MockStore) GetDashboardLayout(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockStore) SaveDashboardLayout(ctx context.Context, userID string, layoutJSON string) error {
	args := m.Called(ctx, userID, layoutJSON)
	return args.Error(0)
}

// Users
func (m *MockStore) CreateUser(ctx context.Context, user *configv1.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockStore) GetUser(ctx context.Context, id string) (*configv1.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.User), args.Error(1)
}

func (m *MockStore) ListUsers(ctx context.Context) ([]*configv1.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.User), args.Error(1)
}

func (m *MockStore) UpdateUser(ctx context.Context, user *configv1.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockStore) DeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Profiles
func (m *MockStore) ListProfiles(ctx context.Context) ([]*configv1.ProfileDefinition, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.ProfileDefinition), args.Error(1)
}

func (m *MockStore) GetProfile(ctx context.Context, name string) (*configv1.ProfileDefinition, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.ProfileDefinition), args.Error(1)
}

func (m *MockStore) SaveProfile(ctx context.Context, profile *configv1.ProfileDefinition) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockStore) DeleteProfile(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// Service Collections
func (m *MockStore) ListServiceCollections(ctx context.Context) ([]*configv1.Collection, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.Collection), args.Error(1)
}

func (m *MockStore) GetServiceCollection(ctx context.Context, name string) (*configv1.Collection, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.Collection), args.Error(1)
}

func (m *MockStore) SaveServiceCollection(ctx context.Context, collection *configv1.Collection) error {
	args := m.Called(ctx, collection)
	return args.Error(0)
}

func (m *MockStore) DeleteServiceCollection(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// Tokens
func (m *MockStore) SaveToken(ctx context.Context, token *configv1.UserToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockStore) GetToken(ctx context.Context, userID, serviceID string) (*configv1.UserToken, error) {
	args := m.Called(ctx, userID, serviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.UserToken), args.Error(1)
}

func (m *MockStore) DeleteToken(ctx context.Context, userID, serviceID string) error {
	args := m.Called(ctx, userID, serviceID)
	return args.Error(0)
}

// Credentials
func (m *MockStore) ListCredentials(ctx context.Context) ([]*configv1.Credential, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.Credential), args.Error(1)
}

func (m *MockStore) GetCredential(ctx context.Context, id string) (*configv1.Credential, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.Credential), args.Error(1)
}

func (m *MockStore) SaveCredential(ctx context.Context, cred *configv1.Credential) error {
	args := m.Called(ctx, cred)
	return args.Error(0)
}

func (m *MockStore) DeleteCredential(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStore) SaveGlobalSettings(ctx context.Context, settings *configv1.GlobalSettings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

func (m *MockStore) HasConfigSources() bool {
	return true
}

func TestInitializeDatabase_Empty(t *testing.T) {
	mockStore := new(MockStore)
	app := &Application{}

	mockStore.On("ListServices", mock.Anything).Return(([]*configv1.UpstreamServiceConfig)(nil), nil)
	mockStore.On("GetGlobalSettings", mock.Anything).Return((*configv1.GlobalSettings)(nil), nil)
	mockStore.On("SaveGlobalSettings", mock.Anything, mock.Anything).Return(nil)
	mockStore.On("SaveService", mock.Anything, mock.Anything).Return(nil)
	// Admin User Init expectations
	mockStore.On("ListUsers", mock.Anything).Return(([]*configv1.User)(nil), nil)
	mockStore.On("CreateUser", mock.Anything, mock.Anything).Return(nil)

	err := app.initializeDatabase(context.Background(), mockStore)
	assert.NoError(t, err)

	mockStore.AssertExpectations(t)
}

func TestInitializeDatabase_AlreadyInitialized(t *testing.T) {
	mockStore := new(MockStore)
	app := &Application{}

	mockStore.On("ListServices", mock.Anything).Return([]*configv1.UpstreamServiceConfig{{}}, nil)

	err := app.initializeDatabase(context.Background(), mockStore)
	assert.NoError(t, err)

	mockStore.AssertNotCalled(t, "SaveGlobalSettings")
	mockStore.AssertNotCalled(t, "SaveService")
}

func TestInitializeDatabase_NotStorage(t *testing.T) {
	simpleMock := new(MockSimpleStore)
	app := &Application{}

	simpleMock.On("Load", mock.Anything).Return(&configv1.McpAnyServerConfig{}, nil)

	err := app.initializeDatabase(context.Background(), simpleMock)
	assert.NoError(t, err)
}

type MockSimpleStore struct {
	mock.Mock
}

func (m *MockSimpleStore) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.McpAnyServerConfig), args.Error(1)
}

func (m *MockSimpleStore) Watch(ctx context.Context) (<-chan *configv1.McpAnyServerConfig, error) {
	return nil, nil
}

func (m *MockSimpleStore) HasConfigSources() bool {
	return true
}

func TestInitializeDatabase_Errors(t *testing.T) {
	t.Run("Store Load Error", func(t *testing.T) {
		mockSimpleStore := new(MockSimpleStore)
		app := &Application{}

		mockSimpleStore.On("Load", mock.Anything).Return((*configv1.McpAnyServerConfig)(nil), errors.New("load error"))

		err := app.initializeDatabase(context.Background(), mockSimpleStore)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "load error")
	})

	t.Run("Storage ListServices Error", func(t *testing.T) {
		mockStore := new(MockStore)
		app := &Application{}

		mockStore.On("ListServices", mock.Anything).Return(([]*configv1.UpstreamServiceConfig)(nil), errors.New("list services error"))

		err := app.initializeDatabase(context.Background(), mockStore)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "list services error")
	})

	t.Run("Storage SaveGlobalSettings Error", func(t *testing.T) {
		mockStore := new(MockStore)
		app := &Application{}

		mockStore.On("ListServices", mock.Anything).Return(([]*configv1.UpstreamServiceConfig)(nil), nil)
		mockStore.On("GetGlobalSettings", mock.Anything).Return((*configv1.GlobalSettings)(nil), nil)
		mockStore.On("SaveGlobalSettings", mock.Anything, mock.Anything).Return(errors.New("save global error"))

		err := app.initializeDatabase(context.Background(), mockStore)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save default global settings")
	})

	t.Run("Storage SaveService Error", func(t *testing.T) {
		mockStore := new(MockStore)
		app := &Application{}

		mockStore.On("ListServices", mock.Anything).Return(([]*configv1.UpstreamServiceConfig)(nil), nil)
		mockStore.On("GetGlobalSettings", mock.Anything).Return((*configv1.GlobalSettings)(nil), nil)
		mockStore.On("SaveGlobalSettings", mock.Anything, mock.Anything).Return(nil)
		mockStore.On("SaveService", mock.Anything, mock.Anything).Return(errors.New("save service error"))

		err := app.initializeDatabase(context.Background(), mockStore)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save default weather service")
	})
}

func TestInitializeAdminUser_RandomPassword(t *testing.T) {
	mockStore := new(MockStore)
	app := &Application{}

	// Mocking empty users list
	mockStore.On("ListUsers", mock.Anything).Return(([]*configv1.User)(nil), nil)

	// Capture the user passed to CreateUser
	var capturedUser *configv1.User
	mockStore.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *configv1.User) bool {
		capturedUser = u
		return true
	})).Return(nil)

	// Ensure environment variables are unset for this test
	t.Setenv("MCPANY_ADMIN_INIT_PASSWORD", "")

	err := app.initializeAdminUser(context.Background(), mockStore)
	assert.NoError(t, err)

	assert.NotNil(t, capturedUser)
	assert.Equal(t, "admin", capturedUser.GetId())

	hash := capturedUser.GetAuthentication().GetBasicAuth().GetPasswordHash()
	assert.NotEmpty(t, hash)

	// Check that the password is NOT "password"
	// passhash.CheckPassword returns true if match
	assert.False(t, passhash.CheckPassword("password", hash), "Randomly generated password should not be 'password'")
}
