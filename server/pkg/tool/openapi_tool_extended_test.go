package tool_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestOpenAPITool_Execute_Extended(t *testing.T) {
	t.Run("POST with Input Template", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assert.Equal(t, `{"name": "test"}`, string(body))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		toolProto := &v1.Tool{}
		toolProto.SetName("testTool")
		mockClient := &mockHTTPClient{
			doFunc: server.Client().Do,
		}

		callDef := &configv1.OpenAPICallDefinition{
			InputTransformer: &configv1.InputTransformer{
				Template: proto.String(`{"name": "{{name}}"}`),
			},
		}

		openAPITool := tool.NewOpenAPITool(toolProto, mockClient, nil, "POST", server.URL, nil, callDef)

		inputs := json.RawMessage(`{"name": "test"}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err := openAPITool.Execute(context.Background(), req)
		require.NoError(t, err)
	})

	t.Run("Output Transformer Template", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data": "success"}`))
		}))
		defer server.Close()

		toolProto := &v1.Tool{}
		toolProto.SetName("testTool")
		mockClient := &mockHTTPClient{
			doFunc: server.Client().Do,
		}

		format := configv1.OutputTransformer_JSON
		callDef := &configv1.OpenAPICallDefinition{
			OutputTransformer: &configv1.OutputTransformer{
				Format:   &format,
				Template: proto.String(`Result: {{data}}`),
				ExtractionRules: map[string]string{
					"data": "{.data}",
				},
			},
		}

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

		toolProto := &v1.Tool{}
		toolProto.SetName("testTool")
		mockClient := &mockHTTPClient{
			doFunc: targetServer.Client().Do,
		}

		callDef := &configv1.OpenAPICallDefinition{
			InputTransformer: &configv1.InputTransformer{
				Webhook: &configv1.WebhookConfig{
					Url: webhookServer.URL,
				},
			},
		}

		openAPITool := tool.NewOpenAPITool(toolProto, mockClient, nil, "POST", targetServer.URL, nil, callDef)

		inputs := json.RawMessage(`{}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err := openAPITool.Execute(context.Background(), req)
		require.NoError(t, err)
	})
}
