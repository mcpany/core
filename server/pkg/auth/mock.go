// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import "net/http"

// MockUpstreamAuthenticator represents the public MockUpstreamAuthenticator entity.
//
// Summary: Defines the structured data model representing a upstream authenticator.
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
type MockUpstreamAuthenticator struct {
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
func (m *MockUpstreamAuthenticator) Authenticate(req *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}
