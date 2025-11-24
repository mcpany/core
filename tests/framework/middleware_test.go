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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMiddlewareHelpers(t *testing.T) {
	t.Run("getUpstreamMetrics", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/metrics", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]int64{"counter": 42})
		}))
		defer server.Close()

		metrics := getUpstreamMetrics(t, server.Listener.Addr().String())
		require.Equal(t, int64(42), metrics["counter"])
	})

	t.Run("callTool", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)

			require.Equal(t, "tools/call", body["method"])
			params, ok := body["params"].(map[string]interface{})
			require.True(t, ok)
			require.Equal(t, "test-tool", params["name"])

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": "1", "result": "ok"})
		}))
		defer server.Close()

		callTool(t, server.URL, "test-tool")
	})
}
