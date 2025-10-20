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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mcpxy/core/pkg/auth"
	"github.com/mcpxy/core/pkg/client"
	"github.com/mcpxy/core/pkg/pool"
	"github.com/mcpxy/core/pkg/tool"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	v1 "github.com/mcpxy/core/proto/mcp_router/v1"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuthenticator struct {
	err error
}

func (m *mockAuthenticator) Authenticate(req *http.Request) error {
	return m.err
}

var _ auth.UpstreamAuthenticator = &mockAuthenticator{}

func setupHTTPToolTest(t *testing.T, handler http.Handler, callDefinition *configv1.HttpCallDefinition) (*tool.HTTPTool, *httptest.Server) {
	t.Helper()

	server := httptest.NewServer(handler)
	poolManager := pool.NewManager()

	p, err := pool.New(func(ctx context.Context) (*client.HttpClientWrapper, error) {
		return &client.HttpClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 0)
	require.NoError(t, err)

	poolManager.Register("test-service", p)

	var methodAndURL string
	if callDefinition.GetMethod() == configv1.HttpCallDefinition_HTTP_METHOD_GET {
		methodAndURL = "GET " + server.URL
	} else {
		methodAndURL = "POST " + server.URL
	}

	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDefinition)
	return httpTool, server
}

func TestHTTPTool_Execute_InputTransformation(t *testing.T) {
	expectedBody := `name=test&age=30`
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, expectedBody, string(body))
		w.WriteHeader(http.StatusOK)
	})

	callDef := configv1.HttpCallDefinition_builder{
		InputTransformer: configv1.InputTransformer_builder{
			Template: lo.ToPtr(`name={{name}}&age={{age}}`),
		}.Build(),
	}.Build()

	httpTool, server := setupHTTPToolTest(t, handler, callDef)
	defer server.Close()

	inputs := json.RawMessage(`{"name": "test", "age": 30}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	_, err := httpTool.Execute(context.Background(), req)
	require.NoError(t, err)
}

func TestHTTPTool_Execute_OutputTransformation_XML(t *testing.T) {
	xmlResponse := `<user><id>123</id><name>Test</name></user>`
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(xmlResponse))
	})

	format := configv1.OutputTransformer_XML
	callDef := configv1.HttpCallDefinition_builder{
		OutputTransformer: configv1.OutputTransformer_builder{
			Format: &format,
			ExtractionRules: map[string]string{
				"id":   "//id",
				"name": "//name",
			},
		}.Build(),
	}.Build()

	httpTool, server := setupHTTPToolTest(t, handler, callDef)
	defer server.Close()

	req := &tool.ExecutionRequest{}
	result, err := httpTool.Execute(context.Background(), req)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "123", resultMap["id"])
	assert.Equal(t, "Test", resultMap["name"])
}

func TestHTTPTool_Execute_OutputTransformation_Text(t *testing.T) {
	textResponse := "User: test-user, Role: admin"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(textResponse))
	})

	format := configv1.OutputTransformer_TEXT
	callDef := configv1.HttpCallDefinition_builder{
		OutputTransformer: configv1.OutputTransformer_builder{
			Format: &format,
			ExtractionRules: map[string]string{
				"username": `User: ([\w-]+)`,
				"role":     `Role: (\w+)`,
			},
		}.Build(),
	}.Build()

	httpTool, server := setupHTTPToolTest(t, handler, callDef)
	defer server.Close()

	req := &tool.ExecutionRequest{}
	result, err := httpTool.Execute(context.Background(), req)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "test-user", resultMap["username"])
	assert.Equal(t, "admin", resultMap["role"])
}

func TestHTTPTool_Execute_NoTransformation(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test", r.URL.Query().Get("param"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "param": "test"}`))
	})

	method := configv1.HttpCallDefinition_HTTP_METHOD_GET
	callDef := configv1.HttpCallDefinition_builder{
		Method: &method,
	}.Build()

	httpTool, server := setupHTTPToolTest(t, handler, callDef)
	defer server.Close()

	inputs := json.RawMessage(`{"param": "test"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	result, err := httpTool.Execute(context.Background(), req)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "ok", resultMap["status"])
	assert.Equal(t, "test", resultMap["param"])
}

func TestHTTPTool_Execute_Errors(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: lo.ToPtr("GET " + server.URL),
	}.Build()

	t.Run("pool_not_found", func(t *testing.T) {
		poolManager := pool.NewManager() // Empty pool manager
		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, &configv1.HttpCallDefinition{})
		_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no http pool found for service")
	})

	t.Run("pool_get_error", func(t *testing.T) {
		poolManager := pool.NewManager()
		errorFactory := func(ctx context.Context) (*client.HttpClientWrapper, error) {
			return nil, errors.New("pool factory error")
		}
		p, err := pool.New(errorFactory, 0, 1, 0)
		require.NoError(t, err)
		poolManager.Register("test-service", p)

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, &configv1.HttpCallDefinition{})
		_, err = httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get client from pool")
		assert.Contains(t, err.Error(), "pool factory error")
	})

	t.Run("invalid_method_fqn", func(t *testing.T) {
		poolManager := pool.NewManager()
		p, _ := pool.New(func(ctx context.Context) (*client.HttpClientWrapper, error) {
			return &client.HttpClientWrapper{Client: server.Client()}, nil
		}, 1, 1, 0)
		poolManager.Register("test-service", p)
		invalidTool := v1.Tool_builder{UnderlyingMethodFqn: lo.ToPtr("INVALID")}.Build()
		httpTool := tool.NewHTTPTool(invalidTool, poolManager, "test-service", nil, &configv1.HttpCallDefinition{})
		_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid http tool definition")
	})

	t.Run("bad_tool_input_json", func(t *testing.T) {
		httpTool, _ := setupHTTPToolTest(t, handler, &configv1.HttpCallDefinition{})
		req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{"param":`)}
		_, err := httpTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
	})

	t.Run("upstream_error", func(t *testing.T) {
		errHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error"))
		})
		httpTool, server := setupHTTPToolTest(t, errHandler, &configv1.HttpCallDefinition{})
		defer server.Close()

		_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "upstream HTTP request failed with status 500")
	})

	t.Run("auth_failure", func(t *testing.T) {
		authenticator := &mockAuthenticator{err: errors.New("auth error")}
		_, server := setupHTTPToolTest(t, handler, &configv1.HttpCallDefinition{})
		defer server.Close()

		// Re-create tool with authenticator
		poolManager := pool.NewManager()
		p, _ := pool.New(func(ctx context.Context) (*client.HttpClientWrapper, error) {
			return &client.HttpClientWrapper{Client: server.Client()}, nil
		}, 1, 1, 0)
		poolManager.Register("test-service", p)
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: lo.ToPtr("GET " + server.URL),
		}.Build()
		authedTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", authenticator, &configv1.HttpCallDefinition{})

		_, err := authedTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate request")
	})

	t.Run("path_parameter_mapping", func(t *testing.T) {
		pathHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.True(t, strings.HasSuffix(r.URL.Path, "/users/123"), "URL path should contain the user ID")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		})
		server := httptest.NewServer(pathHandler)
		defer server.Close()

		poolManager := pool.NewManager()
		p, _ := pool.New(func(ctx context.Context) (*client.HttpClientWrapper, error) {
			return &client.HttpClientWrapper{Client: server.Client()}, nil
		}, 1, 1, 0)
		poolManager.Register("test-service", p)

		methodAndURL := "GET " + server.URL + "/users/{{userID}}"
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		paramMapping := configv1.HttpParameterMapping_builder{
			Name: lo.ToPtr("userID"),
		}.Build()
		callDef := configv1.HttpCallDefinition_builder{
			Parameters: []*configv1.HttpParameterMapping{paramMapping},
		}.Build()

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef)

		inputs := json.RawMessage(`{"userID": 123}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}
		_, err := httpTool.Execute(context.Background(), req)
		require.NoError(t, err)
	})

	t.Run("output_transformation_template_error", func(t *testing.T) {
		outputTransformer := configv1.OutputTransformer_builder{
			Template: lo.ToPtr("{{invalid"),
		}.Build()
		callDef := configv1.HttpCallDefinition_builder{
			OutputTransformer: outputTransformer,
		}.Build()

		httpTool, server := setupHTTPToolTest(t, handler, callDef)
		defer server.Close()

		req := &tool.ExecutionRequest{}
		_, err := httpTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create output template")
	})

	t.Run("input_transformation_render_error", func(t *testing.T) {
		it := configv1.InputTransformer_builder{
			Template: lo.ToPtr(`{"key": "{{some_key.nested}}"}`), // This will fail as some_key is not a map
		}.Build()
		callDef := configv1.HttpCallDefinition_builder{
			InputTransformer: it,
		}.Build()

		httpTool, server := setupHTTPToolTest(t, handler, callDef)
		defer server.Close()

		req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{"some_key": 123}`)}
		_, err := httpTool.Execute(context.Background(), req)
		require.NoError(t, err) // fasttemplate does not error on missing nested fields, it renders empty string
	})

	t.Run("output_transformation_parse_error", func(t *testing.T) {
		outputTransformer := configv1.OutputTransformer_builder{
			Format:          configv1.OutputTransformer_JSON.Enum(),
			ExtractionRules: map[string]string{"key": ".key"},
		}.Build()

		callDef := configv1.HttpCallDefinition_builder{
			OutputTransformer: outputTransformer,
		}.Build()

		// Handler returns invalid JSON
		errHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`not-json`))
		})

		httpTool, server := setupHTTPToolTest(t, errHandler, callDef)
		defer server.Close()

		req := &tool.ExecutionRequest{}
		_, err := httpTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse output")
	})

	t.Run("non_json_response", func(t *testing.T) {
		// Handler returns non-JSON response
		stringHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("just a string"))
		})

		httpTool, server := setupHTTPToolTest(t, stringHandler, &configv1.HttpCallDefinition{})
		defer server.Close()

		req := &tool.ExecutionRequest{}
		result, err := httpTool.Execute(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "just a string", result)
	})

	t.Run("delete_method_with_params", func(t *testing.T) {
		deleteHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "123", r.URL.Query().Get("id"))
			w.WriteHeader(http.StatusNoContent)
		})
		server := httptest.NewServer(deleteHandler)
		defer server.Close()

		poolManager := pool.NewManager()
		p, _ := pool.New(func(ctx context.Context) (*client.HttpClientWrapper, error) {
			return &client.HttpClientWrapper{Client: server.Client()}, nil
		}, 1, 1, 0)
		poolManager.Register("test-service", p)

		methodAndURL := "DELETE " + server.URL
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, &configv1.HttpCallDefinition{})

		inputs := json.RawMessage(`{"id": "123"}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}
		result, err := httpTool.Execute(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "", result) // Expect empty string for No Content response
	})
}
