// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/tests/integration"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func BuildHTTPEchoServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "http_echo_server", filepath.Join(root, "build/test/bin/http_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterHTTPEchoService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_http_echo"
	integration.RegisterHTTPService(t, registrationClient, serviceID, upstreamEndpoint, "echo", "/echo", http.MethodPost, nil)
}

func BuildHTTPAuthedEchoServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "http_authed_echo_server", filepath.Join(root, "build/test/bin/http_authed_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterHTTPAuthedEchoService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_http_authed_echo"
	secret := &configv1.SecretValue{}
	secret.SetPlainText("test-api-key")
	authConfig := configv1.UpstreamAuthentication_builder{
		ApiKey: configv1.UpstreamAPIKeyAuth_builder{
			HeaderName: proto.String("X-Api-Key"),
			ApiKey:     secret,
		}.Build(),
	}.Build()
	integration.RegisterHTTPService(t, registrationClient, serviceID, upstreamEndpoint, "echo", "/echo", http.MethodPost, authConfig)
}
