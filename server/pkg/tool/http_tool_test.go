// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

type mockAuthenticator struct {
	err error
}

func (m *mockAuthenticator) Authenticate(_ *http.Request) error {
	return m.err
}

var _ auth.UpstreamAuthenticator = &mockAuthenticator{}

func setupHTTPToolTest(t *testing.T, handler http.Handler, callDefinition *configv1.HttpCallDefinition) (*tool.HTTPTool, *httptest.Server) {
	t.Helper()

	server := httptest.NewServer(handler)
	poolManager := pool.NewManager()

	p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 0, true)
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

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDefinition, nil, nil, "")
	return httpTool, server
}

func TestHTTPTool_Execute_InputTransformation(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	xmlResponse := `<user><id>123</id><name>Test</name></user>`
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(xmlResponse))
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
	t.Parallel()

	textResponse := "User: test-user, Role: admin"
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(textResponse))
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
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test", r.URL.Query().Get("param"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok", "param": "test"}`))
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
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: lo.ToPtr("GET " + server.URL),
	}.Build()

	t.Run("pool_not_found", func(t *testing.T) {
		t.Parallel()
		poolManager := pool.NewManager() // Empty pool manager
		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, &configv1.HttpCallDefinition{}, nil, nil, "")
		_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no http pool found for service")
	})

	t.Run("pool_get_error", func(t *testing.T) {
		t.Parallel()
		poolManager := pool.NewManager()
		errorFactory := func(_ context.Context) (*client.HTTPClientWrapper, error) {
			return nil, errors.New("pool factory error")
		}
		p, err := pool.New(errorFactory, 0, 1, 0, true)
		require.NoError(t, err)
		poolManager.Register("test-service", p)

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, &configv1.HttpCallDefinition{}, nil, nil, "")
		_, err = httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get client from pool")
		assert.Contains(t, err.Error(), "pool factory error")
	})

	t.Run("invalid_method_fqn", func(t *testing.T) {
		t.Parallel()
		poolManager := pool.NewManager()
		p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
			return &client.HTTPClientWrapper{Client: server.Client()}, nil
		}, 1, 1, 0, true)
		poolManager.Register("test-service", p)
		invalidTool := v1.Tool_builder{UnderlyingMethodFqn: lo.ToPtr("INVALID")}.Build()
		httpTool := tool.NewHTTPTool(invalidTool, poolManager, "test-service", nil, &configv1.HttpCallDefinition{}, nil, nil, "")
		_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid http tool definition")
	})

	t.Run("bad_tool_input_json", func(t *testing.T) {
		t.Parallel()
		httpTool, _ := setupHTTPToolTest(t, handler, &configv1.HttpCallDefinition{})
		req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{"param":`)}
		_, err := httpTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
	})

	t.Run("upstream_error", func(t *testing.T) {
		t.Parallel()
		errHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("internal error"))
		})
		httpTool, server := setupHTTPToolTest(t, errHandler, &configv1.HttpCallDefinition{})
		defer server.Close()

		_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "upstream HTTP request failed with status 500")
	})

	t.Run("auth_failure", func(t *testing.T) {
		t.Parallel()
		authenticator := &mockAuthenticator{err: errors.New("auth error")}
		_, server := setupHTTPToolTest(t, handler, &configv1.HttpCallDefinition{})
		defer server.Close()

		// Re-create tool with authenticator
		poolManager := pool.NewManager()
		p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
			return &client.HTTPClientWrapper{Client: server.Client()}, nil
		}, 1, 1, 0, true)
		poolManager.Register("test-service", p)
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: lo.ToPtr("GET " + server.URL),
		}.Build()
		authedTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", authenticator, &configv1.HttpCallDefinition{}, nil, nil, "")

		_, err := authedTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate request")
	})

	t.Run("path_parameter_mapping", func(t *testing.T) {
		t.Parallel()
		pathHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.True(t, strings.HasSuffix(r.URL.Path, "/users/123"), "URL path should contain the user ID")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "ok"}`))
		})
		server := httptest.NewServer(pathHandler)
		defer server.Close()

		poolManager := pool.NewManager()
		p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
			return &client.HTTPClientWrapper{Client: server.Client()}, nil
		}, 1, 1, 0, true)
		poolManager.Register("test-service", p)

		methodAndURL := "GET " + server.URL + "/users/{{userID}}"
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		paramMapping := configv1.HttpParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name: proto.String("userID"),
			}.Build(),
		}.Build()
		callDef := configv1.HttpCallDefinition_builder{
			Parameters: []*configv1.HttpParameterMapping{paramMapping},
		}.Build()

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

		inputs := json.RawMessage(`{"userID": 123}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}
		_, err := httpTool.Execute(context.Background(), req)
		require.NoError(t, err)
	})

	t.Run("output_transformation_template_error", func(t *testing.T) {
		t.Parallel()
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
		assert.Contains(t, err.Error(), "failed to parse output template")
	})

	t.Run("input_transformation_render_error", func(t *testing.T) {
		t.Parallel()
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
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing key")
	})

	t.Run("output_transformation_parse_error", func(t *testing.T) {
		t.Parallel()
		outputTransformer := configv1.OutputTransformer_builder{
			Format:          configv1.OutputTransformer_JSON.Enum(),
			ExtractionRules: map[string]string{"key": ".key"},
		}.Build()

		callDef := configv1.HttpCallDefinition_builder{
			OutputTransformer: outputTransformer,
		}.Build()

		// Handler returns invalid JSON
		errHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`not-json`))
		})

		httpTool, server := setupHTTPToolTest(t, errHandler, callDef)
		defer server.Close()

		req := &tool.ExecutionRequest{}
		_, err := httpTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse output")
	})

	t.Run("non_json_response", func(t *testing.T) {
		t.Parallel()
		// Handler returns non-JSON response
		stringHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("just a string"))
		})

		httpTool, server := setupHTTPToolTest(t, stringHandler, &configv1.HttpCallDefinition{})
		defer server.Close()

		req := &tool.ExecutionRequest{}
		result, err := httpTool.Execute(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "just a string", result)
	})

	t.Run("delete_method_with_params", func(t *testing.T) {
		t.Parallel()
		deleteHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "123", r.URL.Query().Get("id"))
			w.WriteHeader(http.StatusNoContent)
		})
		server := httptest.NewServer(deleteHandler)
		defer server.Close()

		poolManager := pool.NewManager()
		p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
			return &client.HTTPClientWrapper{Client: server.Client()}, nil
		}, 1, 1, 0, true)
		poolManager.Register("test-service", p)

		methodAndURL := "DELETE " + server.URL
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, &configv1.HttpCallDefinition{}, nil, nil, "")

		inputs := json.RawMessage(`{"id": "123"}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}
		result, err := httpTool.Execute(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "", result) // Expect empty string for No Content response
	})

	t.Run("optional_path_parameter", func(t *testing.T) {
		t.Parallel()
		pathHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/json", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "ok"}`))
		})
		server := httptest.NewServer(pathHandler)
		defer server.Close()

		poolManager := pool.NewManager()
		p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
			return &client.HTTPClientWrapper{Client: server.Client()}, nil
		}, 1, 1, 0, true)
		poolManager.Register("test-service", p)

		methodAndURL := "GET " + server.URL + "/{{ip}}/json"
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		paramMapping := configv1.HttpParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name: proto.String("ip"),
			}.Build(),
		}.Build()
		callDef := configv1.HttpCallDefinition_builder{
			Parameters: []*configv1.HttpParameterMapping{paramMapping},
		}.Build()

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

		inputs := json.RawMessage(`{}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}
		_, err := httpTool.Execute(context.Background(), req)
		require.NoError(t, err)
	})
}

func TestHTTPTool_Execute_InputTransformation_Webhook(t *testing.T) {
	t.Parallel()

	webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/cloudevents+json")
		responseEvent := `{
			"specversion": "1.0",
			"type": "com.mcpany.tool.transform_input.response",
			"source": "webhook-test",
			"id": "123",
			"data": {"transformed": "input"}
		}`
		w.Write([]byte(responseEvent))
	}))
	defer webhookServer.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"transformed": "input"}`, string(body))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	method := configv1.HttpCallDefinition_HTTP_METHOD_POST
	callDef := &configv1.HttpCallDefinition{
		Method: &method,
		InputTransformer: &configv1.InputTransformer{
			Webhook: &configv1.WebhookConfig{
				Url: webhookServer.URL,
			},
		},
	}

	httpTool, server := setupHTTPToolTest(t, handler, callDef)
	defer server.Close()

	inputs := json.RawMessage(`{}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	_, err := httpTool.Execute(context.Background(), req)
	require.NoError(t, err)
}

func TestHTTPTool_Execute_OutputTransformation_RawBytes(t *testing.T) {
	t.Parallel()

	rawBytesResponse := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(rawBytesResponse)
	})

	format := configv1.OutputTransformer_RAW_BYTES
	callDef := configv1.HttpCallDefinition_builder{
		OutputTransformer: configv1.OutputTransformer_builder{
			Format: &format,
		}.Build(),
	}.Build()

	httpTool, server := setupHTTPToolTest(t, handler, callDef)
	defer server.Close()

	req := &tool.ExecutionRequest{}
	result, err := httpTool.Execute(context.Background(), req)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, rawBytesResponse, resultMap["raw"])
}

func TestHTTPTool_Execute_OutputTransformation_JQ(t *testing.T) {
	t.Parallel()

	jsonResponse := `{"users": [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]}`
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(jsonResponse))
	})

	format := configv1.OutputTransformer_JQ
	callDef := configv1.HttpCallDefinition_builder{
		OutputTransformer: configv1.OutputTransformer_builder{
			Format:  &format,
			JqQuery: lo.ToPtr(".users[].name"),
		}.Build(),
	}.Build()

	httpTool, server := setupHTTPToolTest(t, handler, callDef)
	defer server.Close()

	req := &tool.ExecutionRequest{}
	result, err := httpTool.Execute(context.Background(), req)
	require.NoError(t, err)

	// Result should be a slice of interface{}
	resultList, ok := result.([]any)
	require.True(t, ok)
	assert.Contains(t, resultList, "Alice")
	assert.Contains(t, resultList, "Bob")
}

func TestHTTPTool_Execute_PathParameterEncoding(t *testing.T) {
	t.Parallel()

	pathHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/users/test%2Fuser"
		assert.Equal(t, expectedPath, r.URL.RequestURI(), "URL path should be properly escaped")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	server := httptest.NewServer(pathHandler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 0, true)
	require.NoError(t, err)
	poolManager.Register("test-service", p)

	methodAndURL := "GET " + server.URL + "/users/{{username}}"
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	paramMapping := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("username"),
		}.Build(),
	}.Build()
	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{paramMapping},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	inputs := json.RawMessage(`{"username": "test/user"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	_, err = httpTool.Execute(context.Background(), req)
	require.NoError(t, err)
}

func TestHTTPTool_Execute_WithRetry(t *testing.T) {
	t.Parallel()

	t.Run("retry_succeeds", func(t *testing.T) {
		t.Parallel()
		attempt := 0
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if attempt == 0 {
				attempt++
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "ok"}`))
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		poolManager := pool.NewManager()
		p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
			return &client.HTTPClientWrapper{Client: server.Client()}, nil
		}, 1, 1, 0, true)
		poolManager.Register("test-service", p)

		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: lo.ToPtr("GET " + server.URL),
		}.Build()

		resilience := &configv1.ResilienceConfig{}
		retryPolicy := &configv1.RetryConfig{}
		retryPolicy.SetNumberOfRetries(2)
		retryPolicy.SetBaseBackoff(durationpb.New(0))
		resilience.SetRetryPolicy(retryPolicy)

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, &configv1.HttpCallDefinition{}, resilience, nil, "")
		_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.NoError(t, err)
	})

	t.Run("retry_fails", func(t *testing.T) {
		t.Parallel()
		attempt := 0
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			attempt++
			w.WriteHeader(http.StatusInternalServerError)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		poolManager := pool.NewManager()
		p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
			return &client.HTTPClientWrapper{Client: server.Client()}, nil
		}, 1, 1, 0, true)
		poolManager.Register("test-service", p)

		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: lo.ToPtr("GET " + server.URL),
		}.Build()

		resilience := &configv1.ResilienceConfig{}
		retryPolicy := &configv1.RetryConfig{}
		retryPolicy.SetNumberOfRetries(2)
		retryPolicy.SetBaseBackoff(durationpb.New(0))
		resilience.SetRetryPolicy(retryPolicy)

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, &configv1.HttpCallDefinition{}, resilience, nil, "")
		_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Equal(t, 3, attempt)
	})

	t.Run("non_retriable_error", func(t *testing.T) {
		t.Parallel()
		attempt := 0
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			attempt++
			w.WriteHeader(http.StatusBadRequest)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		poolManager := pool.NewManager()
		p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
			return &client.HTTPClientWrapper{Client: server.Client()}, nil
		}, 1, 1, 0, true)
		poolManager.Register("test-service", p)

		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: lo.ToPtr("GET " + server.URL),
		}.Build()

		resilience := &configv1.ResilienceConfig{}
		retryPolicy := &configv1.RetryConfig{}
		retryPolicy.SetNumberOfRetries(2)
		retryPolicy.SetBaseBackoff(durationpb.New(0))
		resilience.SetRetryPolicy(retryPolicy)

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, &configv1.HttpCallDefinition{}, resilience, nil, "")
		_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Equal(t, 1, attempt)
	})

	t.Run("retry_post_succeeds", func(t *testing.T) {
		t.Parallel()
		attempt := 0
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			assert.Equal(t, `{"key":"value"}`, string(body))

			if attempt == 0 {
				attempt++
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "ok"}`))
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		poolManager := pool.NewManager()
		p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
			return &client.HTTPClientWrapper{Client: server.Client()}, nil
		}, 1, 1, 0, true)
		poolManager.Register("test-service", p)

		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: lo.ToPtr("POST " + server.URL),
		}.Build()

		resilience := &configv1.ResilienceConfig{}
		retryPolicy := &configv1.RetryConfig{}
		retryPolicy.SetNumberOfRetries(2)
		retryPolicy.SetBaseBackoff(durationpb.New(0))
		resilience.SetRetryPolicy(retryPolicy)

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, &configv1.HttpCallDefinition{}, resilience, nil, "")
		_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolInputs: json.RawMessage(`{"key":"value"}`),
		})
		require.NoError(t, err)
	})
}

func TestHTTPTool_Execute_ConsecutiveCalls(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id": "` + id + `"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 0, true)
	require.NoError(t, err)
	poolManager.Register("test-service", p)

	methodAndURL := "GET " + server.URL + "?id={{id}}"
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	paramMapping := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("id"),
		}.Build(),
	}.Build()
	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{paramMapping},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	// Call 1
	inputs1 := json.RawMessage(`{"id": "1"}`)
	req1 := &tool.ExecutionRequest{ToolInputs: inputs1}
	result1, err := httpTool.Execute(context.Background(), req1)
	require.NoError(t, err)
	resultMap1, ok := result1.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "1", resultMap1["id"])

	// Call 2
	inputs2 := json.RawMessage(`{"id": "2"}`)
	req2 := &tool.ExecutionRequest{ToolInputs: inputs2}
	result2, err := httpTool.Execute(context.Background(), req2)
	require.NoError(t, err)
	resultMap2, ok := result2.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "2", resultMap2["id"])
}
