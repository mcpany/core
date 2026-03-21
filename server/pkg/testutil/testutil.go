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

// Summary: NewTestPoolManager creates a new pool.Manager for testing purposes. It initializes a default HTTP connection pool and registers it with the manager. Helper to create a pool manager with a default "test-service" HTTP pool.
//
// Parameters:
//   - t (*testing.T): The t parameter.
//
// Returns:
//   - *pool.Manager: The resulting *pool.Manager.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// Summary: MockAuthenticator is a mock implementation of the auth.UpstreamAuthenticator interface. Mock authenticator for testing upstream requests.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type MockAuthenticator struct {
	AuthenticateFunc func(req *http.Request) error
}

// Summary: Authenticate calls the mock AuthenticateFunc if set, otherwise returns nil. Authenticates a request using the mock function.
//
// Parameters:
//   - req (*http.Request): The req parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (m *MockAuthenticator) Authenticate(req *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}
