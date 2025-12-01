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
	"net/http"
	"net/http/httptest"
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
			natsConfigFile := CreateTempNatsConfigFile(t)

			if tc.name == "disabled config" {
				mcpAny := StartMCPANYServerWithNoHealthCheck(t, "config-loading-"+tc.name, "--config-path", absConfigFile, "--config-path", natsConfigFile)
				// The server should exit quickly because there are no enabled services.
				// We just need to wait for the process to terminate.
				select {
				case <-mcpAny.Process.waitDone:
					// Process exited as expected.
				case <-time.After(10 * time.Second):
					t.Fatal("MCPANY server with only disabled services did not exit as expected.")
				}
				mcpAny.CleanupFunc()
				// Since the server is not running, we cannot check for services.
				// The test passes if the server exits cleanly.
				return
			}

			mcpAny := StartMCPANYServer(t, "config-loading-"+tc.name, "--config-path", absConfigFile, "--config-path", natsConfigFile)
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

func TestRemoteConfigLoading(t *testing.T) {
	root, err := GetProjectRoot()
	require.NoError(t, err)

	// Start a mock HTTP server to serve the remote config file.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(root, "examples/remote-config/remote_config.yaml"))
	}))
	defer server.Close()

	t.Run("fails without flag", func(t *testing.T) {
		t.Setenv("MCPANY_BINARY_PATH", filepath.Join(root, "build/bin/server"))
		mcpAny := StartMCPANYServerWithNoHealthCheck(t, "remote-config-without-flag", "--config-paths", server.URL)

		// The server should exit quickly because remote configs are not allowed by default.
		select {
		case <-mcpAny.Process.waitDone:
			// Process exited as expected.
		case <-time.After(10 * time.Second):
			t.Fatal("MCPANY server with remote config did not exit as expected.")
		}
		mcpAny.CleanupFunc()
	})

	t.Run("succeeds with flag", func(t *testing.T) {
		t.Setenv("MCPANY_BINARY_PATH", filepath.Join(root, "build/bin/server"))
		mcpAny := StartMCPANYServer(t, "remote-config-with-flag", "--config-paths", server.URL, "--allow-remote-config")
		defer mcpAny.CleanupFunc()

		conn, err := grpc.Dial(mcpAny.GrpcRegistrationEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		client := v1.NewRegistrationServiceClient(conn)

		require.Eventually(t, func() bool {
			resp, err := client.ListServices(context.Background(), &v1.ListServicesRequest{})
			require.NoError(t, err)

			for _, service := range resp.GetServices() {
				if service.GetName() == "hello-remote" {
					return true
				}
			}
			return false
		}, 10*time.Second, 500*time.Millisecond, "remote service not loaded")
	})
}
