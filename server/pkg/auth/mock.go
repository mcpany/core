// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import "net/http"

// MockUpstreamAuthenticator is a mock implementation of UpstreamAuthenticator for testing.
//
// Summary. Represents a MockUpstreamAuthenticator.
type MockUpstreamAuthenticator struct {
	AuthenticateFunc func(req *http.Request) error
}

// Authenticate provides authenticate functionality.
//
// Summary: Authenticate.
//
// Parameters.
//   - req: The parameter.
//
// Returns.
//   - result: The result.
func (m *MockUpstreamAuthenticator) Authenticate(req *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}
