// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
// BuildHTTPEchoServer builds and starts an HTTP echo server for testing.
// t is the t.
// Returns the result.
//
// Summary: BuildHTTPEchoServer provides functionality.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// RegisterHTTPEchoService registers the HTTP echo service with the MCP server.
// t is the t.
// registrationClient is the registrationClient.
// upstreamEndpoint is the upstreamEndpoint.
//
// Summary: RegisterHTTPEchoService provides functionality.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// BuildHTTPAuthedEchoServer builds the HTTP authed echo server for testing.
// t is the t.
// Returns the result.
//
// Summary: BuildHTTPAuthedEchoServer provides functionality.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// RegisterHTTPAuthedEchoService registers the HTTP authed echo service with the given registration client.
// t is the t.
// registrationClient is the registrationClient.
// upstreamEndpoint is the upstreamEndpoint.
//
// Summary: RegisterHTTPAuthedEchoService provides functionality.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

func BuildHTTPEchoServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	proc := integration.NewManagedProcess(t, "http_echo_server", integration.MockBinary(t, "http_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterHTTPEchoService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_http_echo"
	integration.RegisterHTTPService(t, registrationClient, serviceID, upstreamEndpoint, "echo", "/echo", http.MethodPost, nil)
}

func BuildHTTPAuthedEchoServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	proc := integration.NewManagedProcess(t, "http_authed_echo_server", integration.MockBinary(t, "http_authed_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

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
