// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/require"
)

// NewTestPoolManager creates a new pool.Manager for testing purposes.
//
// Summary: Initializes a pool manager with a default HTTP connection pool.
//
// Parameters:
//   - t (*testing.T): The testing object.
//
// Returns:
//   - (*pool.Manager): The initialized pool manager.
func NewTestPoolManager(t *testing.T) *pool.Manager {
	t.Helper()
	pm := pool.NewManager()
	httpPool, err := pool.New(
		func(_ context.Context) (*client.HTTPClientWrapper, error) {
			return &client.HTTPClientWrapper{Client: &http.Client{Timeout: 5 * time.Second}}, nil
		},
		1,
		1,
		10,
		1*time.Minute,
		false,
	)
	require.NoError(t, err)
	pm.Register("test-service", httpPool)
	return pm
}

// MockAuthenticator is a mock implementation of the auth.UpstreamAuthenticator interface.
type MockAuthenticator struct {
	AuthenticateFunc func(req *http.Request) error
}

// Authenticate calls the mock AuthenticateFunc if set, otherwise returns nil.
//
// Summary: Executes the mock authentication logic.
//
// Parameters:
//   - req (*http.Request): The HTTP request to authenticate.
//
// Returns:
//   - (error): An error if authentication fails.
func (m *MockAuthenticator) Authenticate(req *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}
