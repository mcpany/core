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
// It initializes a default HTTP connection pool and registers it with the manager.
//
// Summary: Helper to create a pool manager with a default "test-service" HTTP pool.
//
// Parameters: - None.
//   - t: *testing.T. The testing object.
//
// Returns: - None.
//   - *pool.Manager: The initialized pool manager.
//
// Side Effects: - None.
//   - Registers "test-service" in the manager.
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

// MockAuthenticator mockAuthenticator represents a mock authenticator.
//
// Summary: MockAuthenticator represents a mock authenticator.
type MockAuthenticator struct {
	AuthenticateFunc func(req *http.Request) error
}

// Authenticate calls the mock AuthenticateFunc if set, otherwise returns nil.
//
// Summary: Authenticates a request using the mock function.
//
// Parameters: - None.
//   - req: *http.Request. The request to authenticate.
//
// Returns: - None.
//   - error: The error from AuthenticateFunc.
//
// Side Effects: - None.
//   - Invokes the injected AuthenticateFunc.
func (m *MockAuthenticator) Authenticate(req *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}
