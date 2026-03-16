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
// Summary: BuildHTTPEchoServer builds and starts an HTTP echo server for testing.
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
func BuildHTTPEchoServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	proc := integration.NewManagedProcess(t, "http_echo_server", integration.MockBinary(t, "http_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterHTTPEchoService registers the HTTP echo service with the MCP server.
//
// Summary: RegisterHTTPEchoService registers the HTTP echo service with the MCP server.
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
func RegisterHTTPEchoService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_http_echo"
	integration.RegisterHTTPService(t, registrationClient, serviceID, upstreamEndpoint, "echo", "/echo", http.MethodPost, nil)
}

// BuildHTTPAuthedEchoServer builds the HTTP authed echo server for testing.
//
// Summary: BuildHTTPAuthedEchoServer builds the HTTP authed echo server for testing.
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
func BuildHTTPAuthedEchoServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	proc := integration.NewManagedProcess(t, "http_authed_echo_server", integration.MockBinary(t, "http_authed_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterHTTPAuthedEchoService registers the HTTP authed echo service with the given registration client.
//
// Summary: RegisterHTTPAuthedEchoService registers the HTTP authed echo service with the given registration client.
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
