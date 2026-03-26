// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import "net/http"

// MockUpstreamAuthenticator is a mock implementation of UpstreamAuthenticator for testing.
//
// Summary: Mock implementation of the UpstreamAuthenticator interface.
type MockUpstreamAuthenticator struct {
	AuthenticateFunc func(req *http.Request) error
}

// Authenticate executes the mock mock authentication function.
//
// Summary: Executes Authenticate operation.
//
// Parameters:
//   - req: *http.Request. The request object.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (m *MockUpstreamAuthenticator) Authenticate(req *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}
