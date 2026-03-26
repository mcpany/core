// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/util/passhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// MockStore implements storage.Storage for testing
// MockStore implements storage.Storage for testing
// Summary: MockStore
	mock.Mock
}

// Load ...
// Summary: Load
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.McpAnyServerConfig), args.Error(1)
}

// Watch ...
// Summary: Watch
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx)
	return nil, args.Error(1)
}

// ListServices ...
// Summary: ListServices
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.UpstreamServiceConfig), args.Error(1)
}

// GetService ...
// Summary: GetService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.UpstreamServiceConfig), args.Error(1)
}

// SaveService ...
// Summary: SaveService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, svc)
	return args.Error(0)
}

// DeleteService ...
// Summary: DeleteService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, id)
	return args.Error(0)
}

// GetGlobalSettings ...
// Summary: GetGlobalSettings
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.GlobalSettings), args.Error(1)
}

// Secrets
// Secrets
// Summary: ListSecrets
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.Secret), args.Error(1)
}

// GetSecret ...
// Summary: GetSecret
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.Secret), args.Error(1)
}

// SaveSecret ...
// Summary: SaveSecret
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, secret)
	return args.Error(0)
}

// DeleteSecret ...
// Summary: DeleteSecret
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Users
// Users
// Summary: CreateUser
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, user)
	return args.Error(0)
}

// GetUser ...
// Summary: GetUser
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.User), args.Error(1)
}

// ListUsers ...
// Summary: ListUsers
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.User), args.Error(1)
}

// UpdateUser ...
// Summary: UpdateUser
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, user)
	return args.Error(0)
}

// DeleteUser ...
// Summary: DeleteUser
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Profiles
// Profiles
// Summary: ListProfiles
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.ProfileDefinition), args.Error(1)
}

// GetProfile ...
// Summary: GetProfile
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.ProfileDefinition), args.Error(1)
}

// SaveProfile ...
// Summary: SaveProfile
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, profile)
	return args.Error(0)
}

// DeleteProfile ...
// Summary: DeleteProfile
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, name)
	return args.Error(0)
}

// Service Collections
// Service Collections
// Summary: ListServiceCollections
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.Collection), args.Error(1)
}

// GetServiceCollection ...
// Summary: GetServiceCollection
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.Collection), args.Error(1)
}

// SaveServiceCollection ...
// Summary: SaveServiceCollection
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, collection)
	return args.Error(0)
}

// DeleteServiceCollection ...
// Summary: DeleteServiceCollection
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, name)
	return args.Error(0)
}

// Tokens
// Tokens
// Summary: SaveToken
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, token)
	return args.Error(0)
}

// GetToken ...
// Summary: GetToken
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, userID, serviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.UserToken), args.Error(1)
}

// DeleteToken ...
// Summary: DeleteToken
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, userID, serviceID)
	return args.Error(0)
}

// Credentials
// Credentials
// Summary: ListCredentials
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.Credential), args.Error(1)
}

// GetCredential ...
// Summary: GetCredential
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.Credential), args.Error(1)
}

// SaveCredential ...
// Summary: SaveCredential
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, cred)
	return args.Error(0)
}

// DeleteCredential ...
// Summary: DeleteCredential
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Close ...
// Summary: Close
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	return args.Error(0)
}

// SaveGlobalSettings ...
// Summary: SaveGlobalSettings
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, settings)
	return args.Error(0)
}

// HasConfigSources ...
// Summary: HasConfigSources
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return true
}

// Service Templates
// Service Templates
// Summary: ListServiceTemplates
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.ServiceTemplate), args.Error(1)
}

// GetServiceTemplate ...
// Summary: GetServiceTemplate
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.ServiceTemplate), args.Error(1)
}

// SaveServiceTemplate ...
// Summary: SaveServiceTemplate
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, tmpl)
	return args.Error(0)
}

// DeleteServiceTemplate ...
// Summary: DeleteServiceTemplate
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, id)
	return args.Error(0)
}

// SaveLog ...
// Summary: SaveLog
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, entry)
	return args.Error(0)
}

// GetRecentLogs ...
// Summary: GetRecentLogs
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*logging.LogEntry), args.Error(1)
}

// TestInitializeDatabase_Empty ...
// Summary: TestInitializeDatabase_Empty
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockStore := new(MockStore)
	app := &Application{}

	mockStore.On("ListServices", mock.Anything).Return(([]*configv1.UpstreamServiceConfig)(nil), nil)
	mockStore.On("GetGlobalSettings", mock.Anything).Return((*configv1.GlobalSettings)(nil), nil)
	mockStore.On("SaveGlobalSettings", mock.Anything, mock.Anything).Return(nil)
	mockStore.On("SaveService", mock.Anything, mock.Anything).Return(nil)
	// Template Init expectations
	// Template Init expectations
	mockStore.On("ListServiceTemplates", mock.Anything).Return(([]*configv1.ServiceTemplate)(nil), nil)
	mockStore.On("SaveServiceTemplate", mock.Anything, mock.Anything).Return(nil)
	// Collection Init expectations
	mockStore.On("ListServiceCollections", mock.Anything).Return(([]*configv1.Collection)(nil), nil)
	mockStore.On("SaveServiceCollection", mock.Anything, mock.Anything).Return(nil)
	// Admin User Init expectations
	mockStore.On("ListUsers", mock.Anything).Return(([]*configv1.User)(nil), nil)
	mockStore.On("CreateUser", mock.Anything, mock.Anything).Return(nil)

	err := app.initializeDatabase(context.Background(), mockStore, nil)
	assert.NoError(t, err)

	mockStore.AssertExpectations(t)
}

// TestInitializeDatabase_AlreadyInitialized ...
// Summary: TestInitializeDatabase_AlreadyInitialized
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockStore := new(MockStore)
	app := &Application{}

	mockStore.On("ListServices", mock.Anything).Return([]*configv1.UpstreamServiceConfig{{}}, nil)

	err := app.initializeDatabase(context.Background(), mockStore, nil)
	assert.NoError(t, err)

	mockStore.AssertNotCalled(t, "SaveGlobalSettings")
	mockStore.AssertNotCalled(t, "SaveService")
}

// TestInitializeDatabase_SkipsWhenConfigProvidesGlobalSettings ...
// Summary: TestInitializeDatabase_SkipsWhenConfigProvidesGlobalSettings
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockStore := new(MockStore)
	app := &Application{}

	cfg := configv1.McpAnyServerConfig_builder{
		GlobalSettings: configv1.GlobalSettings_builder{
			ApiKey: proto.String("demo-key"),
		}.Build(),
	}.Build()

	err := app.initializeDatabase(context.Background(), mockStore, cfg)
	assert.NoError(t, err)

	mockStore.AssertNotCalled(t, "ListServices")
	mockStore.AssertNotCalled(t, "SaveGlobalSettings")
	mockStore.AssertNotCalled(t, "SaveService")
}

// TestInitializeDatabase_NotStorage ...
// Summary: TestInitializeDatabase_NotStorage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	simpleMock := new(MockSimpleStore)
	app := &Application{}

	simpleMock.On("Load", mock.Anything).Return(&configv1.McpAnyServerConfig{}, nil)

	err := app.initializeDatabase(context.Background(), simpleMock, nil)
	assert.NoError(t, err)
}

// MockSimpleStore ...
// Summary: MockSimpleStore
	mock.Mock
}

// Load ...
// Summary: Load
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.McpAnyServerConfig), args.Error(1)
}

// Watch ...
// Summary: Watch
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}

// HasConfigSources ...
// Summary: HasConfigSources
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return true
}

// TestInitializeDatabase_Errors ...
// Summary: TestInitializeDatabase_Errors
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Run("Store Load Error", func(t *testing.T) {
		mockSimpleStore := new(MockSimpleStore)
		app := &Application{}

		mockSimpleStore.On("Load", mock.Anything).Return((*configv1.McpAnyServerConfig)(nil), errors.New("load error"))

		err := app.initializeDatabase(context.Background(), mockSimpleStore, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "load error")
	})

	t.Run("Storage ListServices Error", func(t *testing.T) {
		mockStore := new(MockStore)
		app := &Application{}

		mockStore.On("ListServices", mock.Anything).Return(([]*configv1.UpstreamServiceConfig)(nil), errors.New("list services error"))

		err := app.initializeDatabase(context.Background(), mockStore, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "list services error")
	})

	t.Run("Storage SaveGlobalSettings Error", func(t *testing.T) {
		mockStore := new(MockStore)
		app := &Application{}

		mockStore.On("ListServices", mock.Anything).Return(([]*configv1.UpstreamServiceConfig)(nil), nil)
		mockStore.On("GetGlobalSettings", mock.Anything).Return((*configv1.GlobalSettings)(nil), nil)
		mockStore.On("SaveGlobalSettings", mock.Anything, mock.Anything).Return(errors.New("save global error"))

		err := app.initializeDatabase(context.Background(), mockStore, nil)
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

		err := app.initializeDatabase(context.Background(), mockStore, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save default weather service")
	})
}

// TestInitializeAdminUser_RandomPassword ...
// Summary: TestInitializeAdminUser_RandomPassword
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
