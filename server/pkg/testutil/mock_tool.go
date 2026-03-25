// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
// Execute calls the mock ExecuteFunc if set, otherwise returns nil.
//
// Summary: Executes the mock tool logic.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - req: *tool.ExecutionRequest. The tool execution request.
//
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Returns:
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - any: The result from ExecuteFunc.
//   - error: The error from ExecuteFunc.
//
// Side Effects:
//   - Invokes the injected ExecuteFunc.
// Errors:
//   - triggers relevant error states on failure.
// GetCacheConfig returns nil for the mock tool.
//
// Summary: Returns cache configuration (nil for mock).
//
// Returns:
//   - *configv1.CacheConfig: Always nil.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
package testutil

package testutil

func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}
