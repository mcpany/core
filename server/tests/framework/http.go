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

// BuildHTTPEchoServer serves as a public interface for interacting with BuildHTTPEchoServer.
//
// Summary: Build the http echo server appropriately based on current system conditions.
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
func BuildHTTPEchoServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	proc := integration.NewManagedProcess(t, "http_echo_server", integration.MockBinary(t, "http_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterHTTPEchoService serves as a public interface for interacting with RegisterHTTPEchoService.
//
// Summary: Register the http echo service appropriately based on current system conditions.
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
func RegisterHTTPEchoService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_http_echo"
	integration.RegisterHTTPService(t, registrationClient, serviceID, upstreamEndpoint, "echo", "/echo", http.MethodPost, nil)
}

// BuildHTTPAuthedEchoServer serves as a public interface for interacting with BuildHTTPAuthedEchoServer.
//
// Summary: Build the http authed echo server appropriately based on current system conditions.
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
func BuildHTTPAuthedEchoServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	proc := integration.NewManagedProcess(t, "http_authed_echo_server", integration.MockBinary(t, "http_authed_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterHTTPAuthedEchoService serves as a public interface for interacting with RegisterHTTPAuthedEchoService.
//
// Summary: Register the http authed echo service appropriately based on current system conditions.
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
