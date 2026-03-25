// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import "net/http"

// MockUpstreamAuthenticator is a mock implementation of UpstreamAuthenticator for testing.
//
// Summary: Represents a MockUpstreamAuthenticator.
type MockUpstreamAuthenticator struct {
	AuthenticateFunc func(req *http.Request) error
}

// Authenticate executes the mock mock authentication function. req is the request object. Returns an error if the operation fails.
//
// Parameters: - None.
//   - req (*http.Request): The request object.
//
// Returns: - None.
//   - error: An error if the operation fails.
//
// Errors: - None.
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes Authenticate operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (m *MockUpstreamAuthenticator) Authenticate(req *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}
