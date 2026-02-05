// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"net/http"
	"testing"

	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTestPoolManager(t *testing.T) {
	pm := NewTestPoolManager(t)
	require.NotNil(t, pm)

	// Verify we can get the registered service
	p, ok := pool.Get[*client.HTTPClientWrapper](pm, "test-service")
	assert.True(t, ok)
	assert.NotNil(t, p)
}

func TestMockAuthenticator(t *testing.T) {
	t.Run("AuthenticateFunc set", func(t *testing.T) {
		called := false
		auth := &MockAuthenticator{
			AuthenticateFunc: func(req *http.Request) error {
				called = true
				return nil
			},
		}

		req, _ := http.NewRequest("GET", "http://example.com", nil)
		err := auth.Authenticate(req)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("AuthenticateFunc nil", func(t *testing.T) {
		auth := &MockAuthenticator{}
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		err := auth.Authenticate(req)
		assert.NoError(t, err)
	})
}

func TestMockTool(t *testing.T) {
	t.Run("Tool definition", func(t *testing.T) {
		mt := &MockTool{}
		def := mt.Tool()
		require.NotNil(t, def)
		assert.Equal(t, "mock-tool", def.GetName())
	})

	t.Run("ExecuteFunc set", func(t *testing.T) {
		called := false
		mt := &MockTool{
			ExecuteFunc: func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
				called = true
				return "result", nil
			},
		}

		res, err := mt.Execute(context.Background(), &tool.ExecutionRequest{})
		assert.NoError(t, err)
		assert.Equal(t, "result", res)
		assert.True(t, called)
	})

	t.Run("ExecuteFunc nil", func(t *testing.T) {
		mt := &MockTool{}
		res, err := mt.Execute(context.Background(), &tool.ExecutionRequest{})
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("GetCacheConfig", func(t *testing.T) {
		mt := &MockTool{}
		assert.Nil(t, mt.GetCacheConfig())
	})
}
