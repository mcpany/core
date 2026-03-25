// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
// NewTestPoolManager creates a new pool.Manager for testing purposes.
// It initializes a default HTTP connection pool and registers it with the manager.
//
// Summary: Helper to create a pool manager with a default "test-service" HTTP pool.
//
// Parameters:
//   - t: *testing.T. The testing object.
//
// Returns:
//   - *pool.Manager: The initialized pool manager.
//
// Side Effects:
//   - Registers "test-service" in the manager.
// Errors:
//   - triggers relevant error states on failure.
func NewTestPoolManager(t *testing.T) *pool.Manager {
	t.Helper()
	pm := pool.NewManager()
// Authenticate calls the mock AuthenticateFunc if set, otherwise returns nil.
//
// Summary: Authenticates a request using the mock function.
//
// Parameters:
//   - req: *http.Request. The request to authenticate.
//
// Returns:
//   - error: The error from AuthenticateFunc.
// Errors:
//   - None.
// Side Effects:
//   - None.
//
// Side Effects:
//   - Invokes the injected AuthenticateFunc.
// Errors:
//   - triggers relevant error states on failure.
func (m *MockAuthenticator) Authenticate(req *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}
