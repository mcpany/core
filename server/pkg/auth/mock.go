// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// MockUpstreamAuthenticator is a mock implementation of UpstreamAuthenticator for testing.
//
// Summary: Represents a MockUpstreamAuthenticator.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
package auth

package auth

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
//
// Summary: Executes Authenticate operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
//   - None.
func (m *MockUpstreamAuthenticator) Authenticate(req *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}
