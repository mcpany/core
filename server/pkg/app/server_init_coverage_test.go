// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

		// ListServices returns empty to trigger initialization
		mockStore.On("ListServices", mock.Anything).Return(([]*configv1.UpstreamServiceConfig)(nil), nil)

		// GetGlobalSettings returns nil to proceed
		mockStore.On("GetGlobalSettings", mock.Anything).Return((*configv1.GlobalSettings)(nil), nil)

		// SaveGlobalSettings fails
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

		// SaveService fails
		mockStore.On("SaveService", mock.Anything, mock.Anything).Return(errors.New("save service error"))

		err := app.initializeDatabase(context.Background(), mockStore)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save default weather service")
	})
}
