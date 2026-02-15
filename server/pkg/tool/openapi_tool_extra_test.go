package tool_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestOpenAPITool_ExtraCoverage(t *testing.T) {
	t.Parallel()

	t.Run("POST with Input Template", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assert.Equal(t, `{"name": "test"}`, string(body))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		mockClient := &mockHTTPClient{
			doFunc: server.Client().Do,
		}

		toolProto := &v1.Tool{}
		toolProto.SetName("testTool")

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

		mockClient := &mockHTTPClient{
			doFunc: server.Client().Do,
		}

		toolProto := &v1.Tool{}
		toolProto.SetName("testTool")

		callDef := configv1.OpenAPICallDefinition_builder{
			OutputTransformer: configv1.OutputTransformer_builder{
				Format:   configv1.OutputTransformer_JSON.Enum(),
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

		resMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "Result: success", resMap["result"])
	})

	t.Run("Raw Bytes Output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte{0x01, 0x02})
		}))
		defer server.Close()

		mockClient := &mockHTTPClient{
			doFunc: server.Client().Do,
		}

		toolProto := &v1.Tool{}
		toolProto.SetName("testTool")

		callDef := configv1.OpenAPICallDefinition_builder{
			OutputTransformer: configv1.OutputTransformer_builder{
				Format: configv1.OutputTransformer_RAW_BYTES.Enum(),
			}.Build(),
		}.Build()

		openAPITool := tool.NewOpenAPITool(toolProto, mockClient, nil, "GET", server.URL, nil, callDef)

		req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
		result, err := openAPITool.Execute(context.Background(), req)
		require.NoError(t, err)

		resMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, []byte{0x01, 0x02}, resMap["raw"])
	})
}
