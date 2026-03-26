// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"fmt"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/tests/integration"
)

// BuildEverythingServer buildEverythingServer build everything server.
//
// Summary: BuildEverythingServer build everything server.
//
// Parameters:
//   - t (*testing.T): The t.
//
// Returns:
//   - *integration.ManagedProcess: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func BuildEverythingServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	args := []string{"@modelcontextprotocol/server-everything", "streamableHttp"}
	env := []string{fmt.Sprintf("PORT=%d", port)}
	proc := integration.NewManagedProcess(t, "everything_streamable_server", "npx", args, env)
	proc.IgnoreExitStatusOne = true
	proc.Port = port
	return proc
}

// RegisterEverythingService registerEverythingService register everything service.
//
// Summary: RegisterEverythingService register everything service.
//
// Parameters:
//   - t (*testing.T): The t.
//   - registrationClient (apiv1.RegistrationServiceClient): The registration client.
//   - upstreamEndpoint (string): The upstream endpoint.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func RegisterEverythingService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_everything_server_streamable"
	integration.RegisterStreamableMCPService(t, registrationClient, serviceID, upstreamEndpoint, true, nil)
}
