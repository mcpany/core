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

package tool_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpxy/core/pkg/auth"
	"github.com/mcpxy/core/pkg/client"
	"github.com/mcpxy/core/pkg/tool"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	v1 "github.com/mcpxy/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHttpClient for testing
type mockHttpClient struct {
	client.HttpClient
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	if m.doFunc != nil {
		return m.doFunc(req)
	}
	return nil, errors.New("not implemented")
}

func TestOpenAPITool_Execute(t *testing.T) {
	t.Run("successful execution with path and query params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/users/123", r.URL.Path)
			assert.Equal(t, "test", r.URL.Query().Get("q"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"123", "name":"test"}`))
		}))
		defer server.Close()

		mockClient := &mockHttpClient{
			doFunc: server.Client().Do,
		}

		toolProto := &v1.Tool{}
		toolProto.SetName("getUser")
		parameterDefs := map[string]string{
			"userId": "path",
			"q":      "query",
		}
		openAPITool := tool.NewOpenAPITool(toolProto, mockClient, parameterDefs, "GET", server.URL+"/users/{userId}", nil, &configv1.OpenAPICallDefinition{})

		inputs := json.RawMessage(`{"userId": "123", "q": "test"}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := openAPITool.Execute(context.Background(), req)
		require.NoError(t, err)

		expected := map[string]any{"id": "123", "name": "test"}
		assert.Equal(t, expected, result)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		mockClient := &mockHttpClient{
			doFunc: server.Client().Do,
		}

		toolProto := &v1.Tool{}
		openAPITool := tool.NewOpenAPITool(toolProto, mockClient, nil, "GET", server.URL, nil, &configv1.OpenAPICallDefinition{})

		req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
		_, err := openAPITool.Execute(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("with authentication", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "my-secret-key", r.Header.Get("X-API-Key"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		mockClient := &mockHttpClient{
			doFunc: server.Client().Do,
		}

		toolProto := &v1.Tool{}
		apiKeyAuth := &configv1.UpstreamAPIKeyAuth{}
		apiKeyAuth.SetHeaderName("X-API-Key")
		apiKeyAuth.SetApiKey("my-secret-key")
		authn := &configv1.UpstreamAuthentication{}
		authn.SetApiKey(apiKeyAuth)
		authenticator, err := auth.NewUpstreamAuthenticator(authn)
		require.NoError(t, err)

		openAPITool := tool.NewOpenAPITool(toolProto, mockClient, nil, "GET", server.URL, authenticator, &configv1.OpenAPICallDefinition{})

		req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
		_, err = openAPITool.Execute(context.Background(), req)
		assert.NoError(t, err)
	})
}
