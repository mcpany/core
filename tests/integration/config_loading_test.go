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

package integration

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfigLoading(t *testing.T) {
	testCases := []struct {
		name             string
		configFile       string
		expectedToolName string
	}{
		{
			name:             "json config",
			configFile:       "testdata/config.json",
			expectedToolName: "http-echo-from-json/-/echo",
		},
		{
			name:             "yaml config",
			configFile:       "testdata/config.yaml",
			expectedToolName: "http-echo-from-yaml/-/echo",
		},
		{
			name:             "textproto config",
			configFile:       "testdata/config.textproto",
			expectedToolName: "http-echo-from-textproto/-/echo",
		},
	}

	root, err := GetProjectRoot()
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("MCPXY_BINARY_PATH", filepath.Join(root, "build/bin/server"))
			absConfigFile := filepath.Join(root, "tests", "integration", tc.configFile)
			mcpx := StartMCPXYServer(t, "config-loading-"+tc.name, "--config-paths", absConfigFile)
			defer mcpx.CleanupFunc()

			// Use a client with no timeout for the streaming SSE connection
			sseClient := &http.Client{}

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, "POST", mcpx.HTTPEndpoint, strings.NewReader(`{"jsonrpc":"2.0","method":"tools/list","id":1}`))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			// Explicitly request SSE stream
			req.Header.Set("Accept", "text/event-stream")

			resp, err := sseClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			toolFoundChan := make(chan bool, 1)
			go func() {
				defer close(toolFoundChan)
				scanner := bufio.NewScanner(resp.Body)
				for scanner.Scan() {
					line := scanner.Text()
					if strings.HasPrefix(line, "data:") {
						data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
						var rpcResp struct {
							Result struct {
								Tools []struct {
									Name string `json:"name"`
								} `json:"tools"`
							} `json:"result"`
						}
						if json.Unmarshal([]byte(data), &rpcResp) == nil {
							for _, tool := range rpcResp.Result.Tools {
								if tool.Name == tc.expectedToolName {
									toolFoundChan <- true
									return
								}
							}
						}
					}
				}
			}()

			select {
			case <-toolFoundChan:
				// Test passed
				return
			case <-ctx.Done():
				t.Logf("mcpx server stderr:\n%s", mcpx.Process.StderrString())
				t.Fatalf("timed out waiting for tool %s in SSE stream", tc.expectedToolName)
			}
		})
	}
}
