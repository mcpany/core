// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
)

func TestMCPMetricsIntegration(t *testing.T) {
	// 1. Initialize middleware which registers metrics to DefaultRegisterer
	mw := middleware.NewMCPMetricsMiddleware()

	// 2. Execute a fake request
	mockHandler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "success"},
			},
		}, nil
	}
	handler := mw.Middleware(mockHandler)

	args := map[string]interface{}{"foo": "bar"}
	argsBytes, _ := json.Marshal(args)
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "test_tool",
			Arguments: json.RawMessage(argsBytes),
		},
	}

	_, err := handler(context.Background(), "tools/call", req)
	assert.NoError(t, err)

	// 3. Scrape /metrics endpoint using promhttp
	// promhttp.Handler() uses DefaultGatherer by default, which gathers from DefaultRegisterer
	ts := httptest.NewServer(promhttp.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	bodyStr := string(body)

	// 4. Verify metrics exist in response
	assert.Contains(t, bodyStr, "mcp_requests_total")
	assert.Contains(t, bodyStr, `method="tools/call"`)
	assert.Contains(t, bodyStr, `status="success"`)
}
