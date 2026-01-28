// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHTTPClientRedefined for testing
type mockHTTPClientRedefined struct {
	client.HTTPClient
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClientRedefined) Do(req *http.Request) (*http.Response, error) {
	if m.doFunc != nil {
		return m.doFunc(req)
	}
	return nil, errors.New("not implemented")
}

func TestOpenAPITool_PathTraversal_Security(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the request reaches here with traversal, it's a failure of the security check
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	mockClient := &mockHTTPClientRedefined{
		doFunc: server.Client().Do,
	}

	toolProto := &v1.Tool{}
	toolProto.SetName("getUser")
	parameterDefs := map[string]string{
		"userId": "path",
	}
	// The URL template implies userId is a path segment
	openAPITool := tool.NewOpenAPITool(toolProto, mockClient, parameterDefs, "GET", server.URL+"/users/{{userId}}", nil, &configv1.OpenAPICallDefinition{})

	// Attack payload: try to traverse up
	inputs := json.RawMessage(`{"userId": "../admin"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	// execute the tool
	_, err := openAPITool.Execute(context.Background(), req)

	// We expect an error due to path traversal detection
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal")
}
