// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import "net/http"

// MockUpstreamAuthenticator - Auto-generated documentation.
//
// Summary: MockUpstreamAuthenticator is a mock implementation of UpstreamAuthenticator for testing.
//
// Fields:
//   - Various fields for MockUpstreamAuthenticator.
type MockUpstreamAuthenticator struct {
	AuthenticateFunc func(req *http.Request) error
}

// Authenticate executes the mock mock authentication function. req is the request object. Returns an error if the operation fails.
//
// Parameters:
//   - req (*http.Request): The request object.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None
func (m *MockUpstreamAuthenticator) Authenticate(req *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}
