// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"errors"
	"net/http"
	"testing"

	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/pool"
	"github.com/stretchr/testify/assert"
)

func TestNewTestPoolManager(t *testing.T) {
	pm := NewTestPoolManager(t)
	assert.NotNil(t, pm, "Pool manager should not be nil")

	// Verify that the "test-service" pool is registered.
	p, ok := pool.Get[*client.HttpClientWrapper](pm, "test-service")
	assert.True(t, ok, "Expected 'test-service' pool to be registered")
	assert.NotNil(t, p, "The retrieved pool should not be nil")
}

func TestMockAuthenticator(t *testing.T) {
	t.Run("AuthenticateFunc is nil", func(t *testing.T) {
		mockAuth := &MockAuthenticator{}
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		err := mockAuth.Authenticate(req)
		assert.NoError(t, err, "Expected no error when AuthenticateFunc is nil")
	})

	t.Run("AuthenticateFunc returns no error", func(t *testing.T) {
		mockAuth := &MockAuthenticator{
			AuthenticateFunc: func(req *http.Request) error {
				return nil
			},
		}
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		err := mockAuth.Authenticate(req)
		assert.NoError(t, err, "Expected no error from AuthenticateFunc")
	})

	t.Run("AuthenticateFunc returns an error", func(t *testing.T) {
		expectedErr := errors.New("authentication failed")
		mockAuth := &MockAuthenticator{
			AuthenticateFunc: func(req *http.Request) error {
				return expectedErr
			},
		}
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		err := mockAuth.Authenticate(req)
		assert.Equal(t, expectedErr, err, "Expected an error from AuthenticateFunc")
	})
}
