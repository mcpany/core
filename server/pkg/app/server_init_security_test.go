// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util/passhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInitializeAdminUser_GeneratesRandomPassword(t *testing.T) {
	// Ensure env var is unset
	os.Unsetenv("MCPANY_ADMIN_INIT_PASSWORD")
	// Clean up just in case, though we unset it first
	defer os.Unsetenv("MCPANY_ADMIN_INIT_PASSWORD")

	mockStore := new(MockStore)
	app := &Application{}

	// Setup mocks
	// ListServices returns nil -> triggers init
	mockStore.On("ListServices", mock.Anything).Return(([]*configv1.UpstreamServiceConfig)(nil), nil)
	// GetGlobalSettings returns nil -> triggers init
	mockStore.On("GetGlobalSettings", mock.Anything).Return((*configv1.GlobalSettings)(nil), nil)
	// Save defaults
	mockStore.On("SaveGlobalSettings", mock.Anything, mock.Anything).Return(nil)
	mockStore.On("SaveService", mock.Anything, mock.Anything).Return(nil)

	// ListUsers returns nil -> triggers admin creation
	mockStore.On("ListUsers", mock.Anything).Return(([]*configv1.User)(nil), nil)

	// Capture the created user
	var createdUser *configv1.User
	mockStore.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *configv1.User) bool {
		createdUser = u
		return true
	})).Return(nil)

	err := app.initializeDatabase(context.Background(), mockStore)
	assert.NoError(t, err)

	assert.NotNil(t, createdUser)
	assert.Equal(t, "admin", createdUser.GetId())

	hash := createdUser.Authentication.GetBasicAuth().GetPasswordHash()
	// Verify it is NOT "password"
	// passhash.CheckPassword returns bool
	match := passhash.CheckPassword("password", hash)
	assert.False(t, match, "Password should not be 'password' (default was randomized)")
}

func TestInitializeAdminUser_UsesEnvVar(t *testing.T) {
	os.Setenv("MCPANY_ADMIN_INIT_PASSWORD", "customSecret123")
	defer os.Unsetenv("MCPANY_ADMIN_INIT_PASSWORD")

	mockStore := new(MockStore)
	app := &Application{}

	mockStore.On("ListServices", mock.Anything).Return(([]*configv1.UpstreamServiceConfig)(nil), nil)
	mockStore.On("GetGlobalSettings", mock.Anything).Return((*configv1.GlobalSettings)(nil), nil)
	mockStore.On("SaveGlobalSettings", mock.Anything, mock.Anything).Return(nil)
	mockStore.On("SaveService", mock.Anything, mock.Anything).Return(nil)
	mockStore.On("ListUsers", mock.Anything).Return(([]*configv1.User)(nil), nil)

	var createdUser *configv1.User
	mockStore.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *configv1.User) bool {
		createdUser = u
		return true
	})).Return(nil)

	err := app.initializeDatabase(context.Background(), mockStore)
	assert.NoError(t, err)

	assert.NotNil(t, createdUser)
	hash := createdUser.Authentication.GetBasicAuth().GetPasswordHash()

	// Verify it IS "customSecret123"
	match := passhash.CheckPassword("customSecret123", hash)
	assert.True(t, match, "Password should be 'customSecret123'")
}
