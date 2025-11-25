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

package framework

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/tests/integration"
	"github.com/stretchr/testify/require"
)

// TestRunE2ETestUnit verifies the logic of the test runner itself,
// ensuring it calls the provided functions in the correct order.
func TestRunE2ETestUnit(t *testing.T) {
	var steps []string

	testCase := &E2ETestCase{
		Name:                "unit_test_runner",
		UpstreamServiceType: "http",
		BuildUpstream: func(t *testing.T) *integration.ManagedProcess {
			steps = append(steps, "BuildUpstream")
			// Return a dummy process; it won't be started.
			return integration.NewManagedProcess(t, "dummy", "echo", nil, nil)
		},
		StartMCPANYServer: func(t *testing.T, testName string, extraArgs ...string) *integration.MCPANYTestServerInfo {
			steps = append(steps, "StartMCPANYServer")
			// Return dummy info; server won't be started.
			return &integration.MCPANYTestServerInfo{
				CleanupFunc: func() { steps = append(steps, "CleanupFunc") },
			}
		},
		RegisterUpstream: func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
			steps = append(steps, "RegisterUpstream")
		},
		ValidateTool: func(t *testing.T, mcpanyEndpoint string) {
			steps = append(steps, "ValidateTool")
		},
		ValidateMiddlewares: func(t *testing.T, mcpanyEndpoint string, upstreamEndpoint string) {
			steps = append(steps, "ValidateMiddlewares")
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			steps = append(steps, "InvokeAIClient")
		},
		RegistrationMethods: []RegistrationMethod{GRPCRegistration}, // Use a non-file method to simplify the test
	}

	// This is a mock for the top-level test runner. Since we're not running in parallel,
	// we can safely inspect the `steps` slice after it completes.
	t.Run("E2ERunnerLogic", func(t *testing.T) {
		RunE2ETest(t, testCase)
	})

	expectedSteps := []string{
		"BuildUpstream",
		"StartMCPANYServer",
		"RegisterUpstream",
		"ValidateTool",
		"ValidateMiddlewares",
		"InvokeAIClient",
		"CleanupFunc", // From the deferred call
	}

	// The actual cleanup function is called by the test framework after the test completes,
	// so we check the order of the main steps and the presence of the cleanup trigger.
	require.Len(t, steps, len(expectedSteps), "A different number of steps were executed than expected")
	for i, step := range expectedSteps {
		require.Equal(t, step, steps[i], "Step %d was not as expected", i)
	}
}

// TestVerifyMCPClientUnit tests the client verification logic against a mock server.
func TestVerifyMCPClientUnit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A minimal valid JSON-RPC response for the `initialize` method call
		// that the MCP client sends upon connection.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"version":"1.0","displayName":"Test Server","capabilities":{}}}`))
	}))
	defer server.Close()

	// This function will attempt to connect, which is sufficient to test its basic logic.
	// It will fail on subsequent calls (like tools/list), but the initial connection is what we're testing.
	// We wrap this in a function that is expected to panic because the SDK will fail when it can't find the tools/list method.
	// This is an acceptable limitation for a unit test.
	require.Panics(t, func() {
		VerifyMCPClient(t, server.URL)
	})
}

// TestBuildFunctions verifies that the process builders create correctly configured ManagedProcess objects.
func TestBuildFunctions(t *testing.T) {
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	testCases := []struct {
		name          string
		buildFunc     func(t *testing.T) *integration.ManagedProcess
		expectedBin   string
		expectedArgs  []string
		expectNilProc bool
	}{
		{"BuildGRPCWeatherServer", BuildGRPCWeatherServer, "grpc_weather_server", []string{"--port="}, false},
		{"BuildHTTPEchoServer", BuildHTTPEchoServer, "http_echo_server", []string{"--port"}, false},
		{"BuildHTTPAuthedEchoServer", BuildHTTPAuthedEchoServer, "http_authed_echo_server", []string{"--port"}, false},
		{"BuildCachingServer", BuildCachingServer, "http_caching_server", []string{"--port"}, false},
		{"BuildEverythingServer", BuildEverythingServer, "npx", []string{"@modelcontextprotocol/server-everything", "streamableHttp"}, false},
		{"BuildStdioServer", BuildStdioServer, "", nil, true},
		{"BuildStdioDockerServer", BuildStdioDockerServer, "", nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			proc := tc.buildFunc(t)
			if tc.expectNilProc {
				require.Nil(t, proc)
				return
			}
			require.NotNil(t, proc)
			require.NotZero(t, proc.Port)
			if tc.expectedBin != "" {
				require.True(t, strings.Contains(proc.Cmd().Path, tc.expectedBin), "Expected command path '%s' to contain '%s'", proc.Cmd().Path, tc.expectedBin)
			}
			if tc.expectedArgs != nil {
				for _, arg := range tc.expectedArgs {
					found := false
					for _, a := range proc.Cmd().Args {
						if strings.Contains(a, arg) {
							found = true
							break
						}
					}
					require.True(t, found, "Expected args %v to contain '%s'", proc.Cmd().Args, arg)
				}
			}
			// Verify that the binary path is correct relative to the project root.
			if !strings.HasPrefix(proc.Cmd().Path, "/") && proc.Cmd().Path != "npx" {
				absPath := filepath.Join(root, "build", "test", "bin", filepath.Base(proc.Cmd().Path))
				_, err := os.Stat(absPath)
				if os.IsNotExist(err) {
					// Also check examples bin
					absPath = filepath.Join(root, "build", "examples", "bin", filepath.Base(proc.Cmd().Path))
					_, err = os.Stat(absPath)
					require.NoError(t, err, "Binary not found at expected path: %s", absPath)
				}
			}
		})
	}
}

func TestNewGeminiCLI(t *testing.T) {
	cli := NewGeminiCLI(t)
	require.NotNil(t, cli)
}
