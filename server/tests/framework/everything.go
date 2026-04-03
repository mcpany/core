// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"fmt"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/tests/integration"
)

// BuildEverythingServer serves as a public interface for interacting with BuildEverythingServer.
//
// Summary: Build the everything server appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func BuildEverythingServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	args := []string{"@modelcontextprotocol/server-everything", "streamableHttp"}
	env := []string{fmt.Sprintf("PORT=%d", port)}
	proc := integration.NewManagedProcess(t, "everything_streamable_server", "npx", args, env)
	proc.IgnoreExitStatusOne = true
	proc.Port = port
	return proc
}

// RegisterEverythingService serves as a public interface for interacting with RegisterEverythingService.
//
// Summary: Register the everything service appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func RegisterEverythingService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_everything_server_streamable"
	integration.RegisterStreamableMCPService(t, registrationClient, serviceID, upstreamEndpoint, true, nil)
}
