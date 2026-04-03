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

// NewTestPoolManager serves as a public interface for interacting with NewTestPoolManager.
//
// Summary: Constructs and returns an initialized test pool manager ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// MockAuthenticator represents the public MockAuthenticator entity.
//
// Summary: Defines the structured data model representing a authenticator.
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

// Authenticate serves as a public interface for interacting with Authenticate.
//
// Summary: Authenticate the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *MockAuthenticator) Authenticate(req *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}
