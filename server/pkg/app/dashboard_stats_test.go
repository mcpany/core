package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleDashboardTopTools(t *testing.T) {
	// Initialize middleware to ensure metrics are registered
	mw := middleware.NewToolMetricsMiddleware(nil)

	// Execute some tools to generate metrics
	ctx := context.Background()
	req1 := &tool.ExecutionRequest{
		ToolName: "test_tool_1",
	}
	// We need to inject tool info into context for service_id, or ToolMetricsMiddleware handles it?
	// It checks tool.GetFromContext.

	// Increment count for tool 1 (run twice)
	_, _ = mw.Execute(ctx, req1, func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "result", nil
	})
	_, _ = mw.Execute(ctx, req1, func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "result", nil
	})

	// Increment count for tool 2 (run once)
	req2 := &tool.ExecutionRequest{ToolName: "test_tool_2"}
	_, _ = mw.Execute(ctx, req2, func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "result", nil
	})

	// Create app
	app := &Application{}
	handler := app.handleDashboardTopTools()

	// Make request
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/dashboard/top-tools", nil)
	handler.ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)

	var stats []ToolUsageStats
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	require.NoError(t, err)

	// Verify
	// We might have other metrics from other tests running in parallel or previously,
	// so we check if our tools are present with at least the expected count.

	found1 := false
	found2 := false

	for _, s := range stats {
		if s.Name == "test_tool_1" {
			assert.GreaterOrEqual(t, s.Count, int64(2))
			found1 = true
		}
		if s.Name == "test_tool_2" {
			assert.GreaterOrEqual(t, s.Count, int64(1))
			found2 = true
		}
	}

	assert.True(t, found1, "test_tool_1 not found in stats")
	assert.True(t, found2, "test_tool_2 not found in stats")
}
