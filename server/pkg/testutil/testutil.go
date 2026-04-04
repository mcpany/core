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
// Summary: Creates a new pool.Manager for testing purposes.
//
// Parameters:
//   - t (*testing.T): Parameter.
//
// Returns:
//   - *pool.Manager: Return value.
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

// MockAuthenticator is a mock implementation of the auth.UpstreamAuthenticator interface.
//
// Summary: Is a mock implementation of the auth.UpstreamAuthenticator interface.
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

// Authenticate calls the mock AuthenticateFunc if set, otherwise returns nil.
//
// Summary: Calls the mock AuthenticateFunc if set, otherwise returns nil.
//
// Parameters:
//   - req (*http.Request): Parameter.
//
// Returns:
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
//
// Side Effects:
//   - None.

func (m *MockAuthenticator) Authenticate(req *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}
