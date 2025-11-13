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
		name                 string
		configFile           string
		expectedToolName     string
		toolShouldBeLoaded   bool
		expectServerExit     bool
		expectErrorInLogs    string
		isAbsPath            bool
		skipTempNatsConfig   bool
		additionalConfigPath string
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
		{
			name:              "non-existent config file",
			configFile:        "testdata/non_existent_config.yaml",
			expectServerExit:  true,
			expectErrorInLogs: "failed to stat path",
		},
		{
			name:               "unsupported config file extension",
			configFile:         "testdata/config.dat",
			expectServerExit:   true,
			expectErrorInLogs:  "unsupported config file extension",
			skipTempNatsConfig: true,
		},
		{
			name:              "malformed json config",
			configFile:        "testdata/malformed_config.json",
			expectServerExit:  true,
			expectErrorInLogs: "failed to unmarshal config",
		},
		{
			name:              "malformed yaml config",
			configFile:        "testdata/malformed_config.yaml",
			expectServerExit:  true,
			expectErrorInLogs: "failed to unmarshal config",
		},
		{
			name:              "malformed textproto config",
			configFile:        "testdata/malformed_config.textproto",
			expectServerExit:  true,
			expectErrorInLogs: "failed to unmarshal config",
		},
		{
			name:               "empty config file",
			configFile:         "testdata/empty_config.yaml",
			toolShouldBeLoaded: false,
		},
	}

	root, err := GetProjectRoot()
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("MCPANY_BINARY_PATH", filepath.Join(root, "build/bin/server"))
			var absConfigFile string
			if tc.isAbsPath {
				absConfigFile = tc.configFile
			} else {
				absConfigFile = filepath.Join(root, "tests", "integration", tc.configFile)
			}

			args := []string{"--config-path", absConfigFile}
			if !tc.skipTempNatsConfig {
				natsConfigFile := CreateTempNatsConfigFile(t)
				args = append(args, "--config-path", natsConfigFile)
			}
			if tc.additionalConfigPath != "" {
				args = append(args, "--config-path", tc.additionalConfigPath)
			}

			if tc.expectServerExit {
				mcpAny := StartMCPANYServerWithNoHealthCheck(t, "config-loading-"+tc.name, args...)
				select {
				case <-mcpAny.Process.waitDone:
					// Process exited as expected.
				case <-time.After(10 * time.Second):
					t.Fatal("MCPANY server did not exit as expected.")
				}
				mcpAny.CleanupFunc()
				logs := mcpAny.LogFile.String()
				require.Contains(t, logs, tc.expectErrorInLogs, "expected error message not found in logs")
				return
			}

			if tc.name == "disabled config" {
				mcpAny := StartMCPANYServerWithNoHealthCheck(t, "config-loading-"+tc.name, "--config-path", absConfigFile, "--config-path", CreateTempNatsConfigFile(t))
				select {
				case <-mcpAny.Process.waitDone:
				case <-time.After(10 * time.Second):
					t.Fatal("MCPANY server with only disabled services did not exit as expected.")
				}
				mcpAny.CleanupFunc()
				return
			}

			mcpAny := StartMCPANYServer(t, "config-loading-"+tc.name, args...)
			defer mcpAny.CleanupFunc()

			conn, err := grpc.Dial(mcpAny.GrpcRegistrationEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
			require.NoError(t, err)
			defer conn.Close()

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
