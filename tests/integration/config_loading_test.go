/*
 * Copyright 2025 Author(s) of MCP Any
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

package integration

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	v1 "github.com/mcpany/core/proto/api/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestConfigLoading(t *testing.T) {
	testCases := []struct {
		name               string
		configFile         string
		expectedToolName   string
		toolShouldBeLoaded bool
	}{
		{
			name:               "json config",
			configFile:         "testdata/config.json",
			expectedToolName:   "http-echo-from-json",
			toolShouldBeLoaded: true,
		},
		{
			name:               "yaml config",
			configFile:         "testdata/config.yaml",
			expectedToolName:   "http-echo-from-yaml",
			toolShouldBeLoaded: true,
		},
		{
			name:               "textproto config",
			configFile:         "testdata/config.textproto",
			expectedToolName:   "http-echo-from-textproto",
			toolShouldBeLoaded: true,
		},
		{
			name:               "disabled config",
			configFile:         "testdata/disabled_config.yaml",
			expectedToolName:   "disabled-service",
			toolShouldBeLoaded: false,
		},
	}

	root, err := GetProjectRoot()
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("MCPANY_BINARY_PATH", filepath.Join(root, "build/bin/server"))
			absConfigFile := filepath.Join(root, "tests", "integration", tc.configFile)


			mcpAny := StartMCPANYServer(t, "config-loading-"+tc.name, "--config-path", absConfigFile)
			defer mcpAny.CleanupFunc()

			conn, err := grpc.NewClient(mcpAny.GrpcRegistrationEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
			require.NoError(t, err)
			defer func() { _ = conn.Close() }()

			client := v1.NewRegistrationServiceClient(conn)

			require.Eventually(t, func() bool {
				resp, err := client.ListServices(context.Background(), &v1.ListServicesRequest{})
				require.NoError(t, err)

				var serviceFound bool
				for _, service := range resp.GetServices() {
					if service.GetName() == tc.expectedToolName {
						serviceFound = true
						break
					}
				}
				return serviceFound == tc.toolShouldBeLoaded
			}, 10*time.Second, 500*time.Millisecond, "service loading status mismatch")
		})
	}
}
