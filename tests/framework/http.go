/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package framework

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	apiv1 "github.com/mcpxy/core/proto/api/v1"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/mcpxy/core/tests/integration"
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
	authConfig := configv1.UpstreamAuthentication_builder{
		ApiKey: configv1.UpstreamAPIKeyAuth_builder{
			HeaderName: proto.String("X-Api-Key"),
			ApiKey:     proto.String("test-api-key"),
		}.Build(),
	}.Build()
	integration.RegisterHTTPService(t, registrationClient, serviceID, upstreamEndpoint, "echo", "/echo", http.MethodPost, authConfig)
}
