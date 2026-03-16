// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"fmt"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/tests/integration"
)

// BuildEverythingServer builds a server with everything.
//
// Summary: BuildEverythingServer builds a server with everything.
//
// Parameters:
//   - t (*testing.T): The t parameter.
//
// Returns:
//   - *integration.ManagedProcess: The *integration.ManagedProcess result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func BuildEverythingServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	args := []string{"@modelcontextprotocol/server-everything", "streamableHttp"}
	env := []string{fmt.Sprintf("PORT=%d", port)}
	proc := integration.NewManagedProcess(t, "everything_streamable_server", "npx", args, env)
	proc.IgnoreExitStatusOne = true
	proc.Port = port
	return proc
}

// RegisterEverythingService registers everything service.
//
// Summary: RegisterEverythingService registers everything service.
//
// Parameters:
//   - t (*testing.T): The t parameter.
//   - registrationClient (apiv1.RegistrationServiceClient): The registrationClient parameter.
//   - upstreamEndpoint (string): The upstreamEndpoint parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func RegisterEverythingService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_everything_server_streamable"
	integration.RegisterStreamableMCPService(t, registrationClient, serviceID, upstreamEndpoint, true, nil)
}
