// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
)

type mockOpenAPIHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockOpenAPIHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func TestOpenAPITool_ErrorBodyRedaction(t *testing.T) {
	t.Run("Non-JSON stack trace is hidden when debug is off", func(t *testing.T) {
		os.Setenv("MCPANY_DEBUG", "false")
		defer os.Unsetenv("MCPANY_DEBUG")

		toolDef := v1.Tool_builder{Name: proto.String("test")}.Build()
		callDef := &configv1.OpenAPICallDefinition{}

		stackTraceBody := "Internal Server Error\nException at /var/www/html/index.php:42\nStack trace:\n#0 main()"
		client := &mockOpenAPIHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
					Body:       io.NopCloser(bytes.NewBufferString(stackTraceBody)),
				}, nil
			},
		}

		openapiTool := NewOpenAPITool(toolDef, client, nil, "GET", "http://example.com", nil, callDef)

		_, err := openapiTool.Execute(context.Background(), &ExecutionRequest{
			ToolInputs: []byte(`{}`),
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "upstream OpenAPI request failed with status 500")
		assert.Contains(t, err.Error(), "[Body hidden for security. Enable debug mode to view.]")
		assert.NotContains(t, err.Error(), "Exception at")
	})

	t.Run("JSON error is redacted and truncated", func(t *testing.T) {
		os.Setenv("MCPANY_DEBUG", "false")
		defer os.Unsetenv("MCPANY_DEBUG")

		toolDef := v1.Tool_builder{Name: proto.String("test")}.Build()
		callDef := &configv1.OpenAPICallDefinition{}

		jsonBody := `{"error": "bad request", "api_key": "sk-123456789", "details": "`
		for i := 0; i < 300; i++ {
			jsonBody += "a"
		}
		jsonBody += `"}`

		client := &mockOpenAPIHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 400,
					Body:       io.NopCloser(bytes.NewBufferString(jsonBody)),
				}, nil
			},
		}

		openapiTool := NewOpenAPITool(toolDef, client, nil, "GET", "http://example.com", nil, callDef)

		_, err := openapiTool.Execute(context.Background(), &ExecutionRequest{
			ToolInputs: []byte(`{}`),
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "upstream OpenAPI request failed with status 400")
		// "api_key" should be redacted
		assert.Contains(t, err.Error(), `"[REDACTED]"`)
		assert.NotContains(t, err.Error(), "sk-123456789")
		// Body should be truncated
		assert.Contains(t, err.Error(), "... (truncated)")
	})
}
