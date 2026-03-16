// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import "net/http"

// MockUpstreamAuthenticator is a mock implementation of UpstreamAuthenticator for testing.
//
// Summary: MockUpstreamAuthenticator is a mock implementation of UpstreamAuthenticator for testing.
//
// Summary: MockUpstreamAuthenticator is a mock implementation of UpstreamAuthenticator for testing.
type MockUpstreamAuthenticator struct {
	AuthenticateFunc func(req *http.Request) error
// Authenticate executes the mock mock authentication function. req is the request object. Returns an error if the operation fails.
//
// Summary: Authenticate executes the mock mock authentication function. req is the request object. Returns an error if the operation fails.
//
// Parameters:
//   - req (*http.Request): The incoming request payload.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (m *MockUpstreamAuthenticator) Authenticate(req *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}
