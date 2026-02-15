package tool_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// mockHTTPClient for testing
type mockHTTPClient struct {
	client.HTTPClient
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.doFunc != nil {
		return m.doFunc(req)
	}
	return nil, errors.New("not implemented")
}

func TestOpenAPITool_Execute(t *testing.T) {
	t.Parallel()
	t.Run("successful execution with path and query params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/users/123", r.URL.Path)
			assert.Equal(t, "test", r.URL.Query().Get("q"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"123", "name":"test"}`))
		}))
		defer server.Close()

		mockClient := &mockHTTPClient{
			doFunc: server.Client().Do,
		}

		toolProto := v1.Tool_builder{
			Name: proto.String("getUser"),
		}.Build()
		parameterDefs := map[string]string{
			"userId": "path",
			"q":      "query",
		}
		openAPITool := tool.NewOpenAPITool(toolProto, mockClient, parameterDefs, "GET", server.URL+"/users/{{userId}}", nil, configv1.OpenAPICallDefinition_builder{}.Build())

		inputs := json.RawMessage(`{"userId": "123", "q": "test"}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := openAPITool.Execute(context.Background(), req)
		require.NoError(t, err)

		expected := map[string]any{"id": "123", "name": "test"}
		assert.Equal(t, expected, result)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		mockClient := &mockHTTPClient{
			doFunc: server.Client().Do,
		}

		toolProto := v1.Tool_builder{}.Build()
		openAPITool := tool.NewOpenAPITool(toolProto, mockClient, nil, "GET", server.URL, nil, configv1.OpenAPICallDefinition_builder{}.Build())

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

		mockClient := &mockHTTPClient{
			doFunc: server.Client().Do,
		}

		toolProto := v1.Tool_builder{}.Build()
		apiKeyAuth := configv1.APIKeyAuth_builder{
			ParamName: proto.String("X-API-Key"),
			Value: configv1.SecretValue_builder{
				PlainText: proto.String("my-secret-key"),
			}.Build(),
		}.Build()

		authn := configv1.Authentication_builder{
			ApiKey: apiKeyAuth,
		}.Build()
		authenticator, err := auth.NewUpstreamAuthenticator(authn)
		require.NoError(t, err)

		openAPITool := tool.NewOpenAPITool(toolProto, mockClient, nil, "GET", server.URL, authenticator, configv1.OpenAPICallDefinition_builder{}.Build())

		req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
		_, err = openAPITool.Execute(context.Background(), req)
		assert.NoError(t, err)
	})
}

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

func TestOpenAPITool_Execute_Extended(t *testing.T) {
	t.Parallel()
	t.Run("POST with Input Template", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assert.Equal(t, `{"name": "test"}`, string(body))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		toolProto := v1.Tool_builder{
			Name: proto.String("testTool"),
		}.Build()
		mockClient := &mockHTTPClient{
			doFunc: server.Client().Do,
		}

		callDef := configv1.OpenAPICallDefinition_builder{
			InputTransformer: configv1.InputTransformer_builder{
				Template: proto.String(`{"name": "{{name}}"}`),
			}.Build(),
		}.Build()

		openAPITool := tool.NewOpenAPITool(toolProto, mockClient, nil, "POST", server.URL, nil, callDef)

		inputs := json.RawMessage(`{"name": "test"}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err := openAPITool.Execute(context.Background(), req)
		require.NoError(t, err)
	})

	t.Run("Output Transformer Template", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data": "success"}`))
		}))
		defer server.Close()

		toolProto := v1.Tool_builder{
			Name: proto.String("testTool"),
		}.Build()
		mockClient := &mockHTTPClient{
			doFunc: server.Client().Do,
		}

		format := configv1.OutputTransformer_JSON
		callDef := configv1.OpenAPICallDefinition_builder{
			OutputTransformer: configv1.OutputTransformer_builder{
				Format:   &format,
				Template: proto.String(`Result: {{data}}`),
				ExtractionRules: map[string]string{
					"data": "{.data}",
				},
			}.Build(),
		}.Build()

		openAPITool := tool.NewOpenAPITool(toolProto, mockClient, nil, "GET", server.URL, nil, callDef)

		inputs := json.RawMessage(`{}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := openAPITool.Execute(context.Background(), req)
		require.NoError(t, err)

		// Result should be map[string]any{"result": "Result: success"}
		resMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "Result: success", resMap["result"])
	})

	t.Run("Input Transformer via Webhook", func(t *testing.T) {
		webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/cloudevents+json")
			// A simple JSON CloudEvent
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

		targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assert.JSONEq(t, `{"transformed": "input"}`, string(body))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer targetServer.Close()

		toolProto := v1.Tool_builder{
			Name: proto.String("testTool"),
		}.Build()
		mockClient := &mockHTTPClient{
			doFunc: targetServer.Client().Do,
		}

		callDef := configv1.OpenAPICallDefinition_builder{
			InputTransformer: configv1.InputTransformer_builder{
				Webhook: configv1.WebhookConfig_builder{
					Url: webhookServer.URL,
				}.Build(),
			}.Build(),
		}.Build()

		openAPITool := tool.NewOpenAPITool(toolProto, mockClient, nil, "POST", targetServer.URL, nil, callDef)

		inputs := json.RawMessage(`{}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err := openAPITool.Execute(context.Background(), req)
		require.NoError(t, err)
	})
}
