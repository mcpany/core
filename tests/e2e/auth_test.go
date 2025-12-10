// Copyright 2024 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/mcpany/core/tests/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_APIKeyAuthentication(t *testing.T) {
	configContent := `
global_settings:
  api_key: "a-very-secret-and-long-api-key"
`
	mcpanyTestServerInfo := integration.StartMCPANYServerWithConfig(t, "api_key_test", configContent)
	defer mcpanyTestServerInfo.CleanupFunc()

	// Case 1: No API Key
	resp, err := http.Post(mcpanyTestServerInfo.HTTPEndpoint, "application/json", strings.NewReader(`{"jsonrpc":"2.0","method":"tools/list","id":1}`))
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// Case 2: Incorrect API Key
	req, err := http.NewRequest("POST", mcpanyTestServerInfo.HTTPEndpoint, strings.NewReader(`{"jsonrpc":"2.0","method":"tools/list","id":1}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "incorrect-key")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// Case 3: Correct API Key
	req, err = http.NewRequest("POST", mcpanyTestServerInfo.HTTPEndpoint, strings.NewReader(`{"jsonrpc":"2.0","method":"tools/list","id":1}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "a-very-secret-and-long-api-key")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, float64(1), result["id"])
	resp.Body.Close()
}

func TestE2E_NoAPIKey(t *testing.T) {
	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "no_api_key_test")
	defer mcpanyTestServerInfo.CleanupFunc()

	resp, err := http.Post(fmt.Sprintf("%s/mcp", mcpanyTestServerInfo.JSONRPCEndpoint), "application/json", strings.NewReader(`{"jsonrpc":"2.0","method":"tools/list","id":1}`))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}
