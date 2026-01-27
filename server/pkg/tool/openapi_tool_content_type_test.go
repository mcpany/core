// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestOpenAPITool_ContentType_JSON_Template(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, `{"name": "test"}`, string(body))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	toolProto := v1.Tool_builder{}.Build()
	mockClient := &mockHTTPClient{
		doFunc: server.Client().Do,
	}

	callDef := configv1.OpenAPICallDefinition_builder{
		InputTransformer: configv1.InputTransformer_builder{
			Template: proto.String(`{"name": "{{name}}"}`),
		}.Build(),
	}.Build()

	openAPITool := NewOpenAPITool(toolProto, mockClient, nil, "POST", server.URL, nil, callDef)

	inputs := json.RawMessage(`{"name": "test"}`)
	req := &ExecutionRequest{ToolName: "testTool", ToolInputs: inputs}

	_, err := openAPITool.Execute(context.Background(), req)
	require.NoError(t, err)
}
