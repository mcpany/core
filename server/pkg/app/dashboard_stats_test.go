// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleDashboardToolFailures(t *testing.T) {
	// Define counters matching what the middleware uses
	toolsCallTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcpany_tools_call_total",
			Help: "Total number of tool calls",
		},
		[]string{"tool", "service_id", "status", "error_type"},
	)

	// Register it if not already registered (handle panic or error)
	if err := prometheus.Register(toolsCallTotal); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			toolsCallTotal = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			// If it's another error, we might just proceed if it's already there?
			// But for test stability, let's just log.
			t.Logf("Failed to register metric: %v", err)
			// Assuming it exists, we try to retrieve it from DefaultGatherer later,
			// but we need the object to Add values.
			// We can't easily get the object back unless we keep global reference or lookup.
			// For this test, if it fails to register because it exists, we assume `toolsCallTotal`
			// now points to the existing one.
		}
	}

    // Clear existing metrics? Not easily possible with standard Prometheus client without resetting registry.
    // We will use unique tool names for this test to avoid collision with other tests.

    toolA := "test_tool_failures_A"
    toolB := "test_tool_failures_B"
    toolD := "test_tool_failures_D"

	// Tool A: 10 success, 10 error => 50% failure
	toolsCallTotal.WithLabelValues(toolA, "service1", "success", "").Add(10)
	toolsCallTotal.WithLabelValues(toolA, "service1", "error", "some_error").Add(10)

	// Tool B: 90 success, 10 error => 10% failure
	toolsCallTotal.WithLabelValues(toolB, "service1", "success", "").Add(90)
	toolsCallTotal.WithLabelValues(toolB, "service1", "error", "timeout").Add(10)

	// Tool D: 0 success, 5 error => 100% failure
	toolsCallTotal.WithLabelValues(toolD, "service2", "error", "crash").Add(5)

	// Create Request
	req, err := http.NewRequest("GET", "/dashboard/tool-failures", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	app := &Application{} // We don't need dependencies for this handler

	handler := app.handleDashboardToolFailures()
	handler.ServeHTTP(rr, req)

	// Check Response
	assert.Equal(t, http.StatusOK, rr.Code)

	var stats []ToolFailureStats
	err = json.Unmarshal(rr.Body.Bytes(), &stats)
	require.NoError(t, err)

    // Filter stats to include only our test tools
    var myStats []ToolFailureStats
    for _, s := range stats {
        if s.Name == toolA || s.Name == toolB || s.Name == toolD {
            myStats = append(myStats, s)
        }
    }

	// Expect descending order of failure rate: D (100) > A (50) > B (10)
	require.GreaterOrEqual(t, len(myStats), 3)

	assert.Equal(t, toolD, myStats[0].Name)
	assert.Equal(t, 100.0, myStats[0].FailureRate)

	assert.Equal(t, toolA, myStats[1].Name)
	assert.Equal(t, 50.0, myStats[1].FailureRate)

	assert.Equal(t, toolB, myStats[2].Name)
	assert.Equal(t, 10.0, myStats[2].FailureRate)
}
