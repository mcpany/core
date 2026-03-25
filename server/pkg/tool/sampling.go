// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
// Session defines the interface for tools to interact with the client session.
	// ListRoots requests the list of roots from the client.
	//
	// Summary: Requests roots list.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//
	// Returns:
	//   - *mcp.ListRootsResult: The list of roots.
// NewContextWithSession creates a new context with the given Session.
//
// Summary: Injects Session into context.
// GetSession retrieves the Session from the context.
//
// NewContextWithSampler creates a new context with the given Sampler.
//
// Errors:
//   - None.
// Side Effects:
//   - None.
// Summary: Injects Sampler into context.
// GetSampler retrieves the Sampler from the context.
//
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Retrieves Sampler from context.
//
// Deprecated: Use GetSession instead.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Parameters:
//   - ctx: context.Context. The context.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Returns:
//   - Sampler: The sampler if found.
//   - bool: True if the sampler exists.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func GetSampler(ctx context.Context) (Sampler, bool) {
	return GetSession(ctx)
}
