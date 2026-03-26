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

// BuildHTTPEchoServer buildHTTPEchoServer build http echo server.
//
// Summary: BuildHTTPEchoServer build http echo server.
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
func BuildHTTPEchoServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	proc := integration.NewManagedProcess(t, "http_echo_server", integration.MockBinary(t, "http_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterHTTPEchoService registerHTTPEchoService register http echo service.
//
// Summary: RegisterHTTPEchoService register http echo service.
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
func RegisterHTTPEchoService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_http_echo"
	integration.RegisterHTTPService(t, registrationClient, serviceID, upstreamEndpoint, "echo", "/echo", http.MethodPost, nil)
}

// BuildHTTPAuthedEchoServer buildHTTPAuthedEchoServer build http authed echo server.
//
// Summary: BuildHTTPAuthedEchoServer build http authed echo server.
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
func BuildHTTPAuthedEchoServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	proc := integration.NewManagedProcess(t, "http_authed_echo_server", integration.MockBinary(t, "http_authed_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterHTTPAuthedEchoService registerHTTPAuthedEchoService register http authed echo service.
//
// Summary: RegisterHTTPAuthedEchoService register http authed echo service.
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
