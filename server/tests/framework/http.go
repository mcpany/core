// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"fmt"
	"net/http"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/tests/integration"
	"google.golang.org/protobuf/proto"
)

// BuildHTTPEchoServer builds and starts an HTTP echo server for testing.
//
// t is the t.
//
// Returns the result.
//
// Summary: Implements BuildHTTPEchoServer for the system.
//
// Parameters:
//   - t: Contextual argument for BuildHTTPEchoServer.
//
// Returns:
//   - *integration.ManagedProcess: The computed output or reference.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies local state or underlying systems.
func BuildHTTPEchoServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	proc := integration.NewManagedProcess(t, "http_echo_server", integration.MockBinary(t, "http_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterHTTPEchoService registers the HTTP echo service with the MCP server.
//
// t is the t.
// registrationClient is the registrationClient.
// upstreamEndpoint is the upstreamEndpoint.
//
// Summary: Implements RegisterHTTPEchoService for the system.
//
// Parameters:
//   - t: Contextual argument for RegisterHTTPEchoService.
//   - registrationClient: Contextual argument for RegisterHTTPEchoService.
//   - upstreamEndpoint: Contextual argument for RegisterHTTPEchoService.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies local state or underlying systems.
func RegisterHTTPEchoService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_http_echo"
	integration.RegisterHTTPService(t, registrationClient, serviceID, upstreamEndpoint, "echo", "/echo", http.MethodPost, nil)
}

// BuildHTTPAuthedEchoServer builds the HTTP authed echo server for testing.
//
// t is the t.
//
// Returns the result.
//
// Summary: Implements BuildHTTPAuthedEchoServer for the system.
//
// Parameters:
//   - t: Contextual argument for BuildHTTPAuthedEchoServer.
//
// Returns:
//   - *integration.ManagedProcess: The computed output or reference.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies local state or underlying systems.
func BuildHTTPAuthedEchoServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	proc := integration.NewManagedProcess(t, "http_authed_echo_server", integration.MockBinary(t, "http_authed_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterHTTPAuthedEchoService registers the HTTP authed echo service with the given registration client.
//
// t is the t.
// registrationClient is the registrationClient.
// upstreamEndpoint is the upstreamEndpoint.
//
// Summary: Implements RegisterHTTPAuthedEchoService for the system.
//
// Parameters:
//   - t: Contextual argument for RegisterHTTPAuthedEchoService.
//   - registrationClient: Contextual argument for RegisterHTTPAuthedEchoService.
//   - upstreamEndpoint: Contextual argument for RegisterHTTPAuthedEchoService.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies local state or underlying systems.
func RegisterHTTPAuthedEchoService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_http_authed_echo"
	secret := configv1.SecretValue_builder{
		PlainText: proto.String("test-api-key"),
	}.Build()
	authConfig := configv1.Authentication_builder{
		ApiKey: configv1.APIKeyAuth_builder{
			ParamName: proto.String("X-Api-Key"),
			Value:     secret,
		}.Build(),
	}.Build()
	integration.RegisterHTTPService(t, registrationClient, serviceID, upstreamEndpoint, "echo", "/echo", http.MethodPost, authConfig)
}
