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

	// Use a local registry for isolation
	registry := prometheus.NewRegistry()
	require.NoError(t, registry.Register(toolsCallTotal))

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
	app := &Application{
		MetricsGatherer: registry,
	}

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

func TestHandleDashboardToolUsage(t *testing.T) {
	// Define counters matching what the middleware uses
	toolsCallTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcpany_tools_call_total",
			Help: "Total number of tool calls",
		},
		[]string{"tool", "service_id", "status", "error_type"},
	)

	// Use a local registry for isolation
	registry := prometheus.NewRegistry()
	require.NoError(t, registry.Register(toolsCallTotal))

	toolA := "tool_A"
	toolB := "tool_B"

	// Tool A: 20 success, 5 error
	toolsCallTotal.WithLabelValues(toolA, "svc1", "success", "").Add(20)
	toolsCallTotal.WithLabelValues(toolA, "svc1", "error", "err1").Add(5)

	// Tool B: 10 success
	toolsCallTotal.WithLabelValues(toolB, "svc2", "success", "").Add(10)

	// Create Request
	req, err := http.NewRequest("GET", "/dashboard/tool-usage", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	app := &Application{
		MetricsGatherer: registry,
	}

	handler := app.handleDashboardToolUsage()
	handler.ServeHTTP(rr, req)

	// Check Response
	assert.Equal(t, http.StatusOK, rr.Code)

	var stats []ToolAnalytics
	err = json.Unmarshal(rr.Body.Bytes(), &stats)
	require.NoError(t, err)

	// Filter stats to include only our test tools
	var myStats []ToolAnalytics
	for _, s := range stats {
		if s.Name == toolA || s.Name == toolB {
			myStats = append(myStats, s)
		}
	}

	require.Equal(t, 2, len(myStats))

	// toolA comes first (alphabetical)
	assert.Equal(t, toolA, myStats[0].Name)
	assert.Equal(t, int64(25), myStats[0].TotalCalls)
	assert.Equal(t, int64(20), myStats[0].SuccessCount)
	assert.Equal(t, int64(5), myStats[0].ErrorCount)
	assert.InDelta(t, 20.0, myStats[0].FailureRate, 0.01)

	assert.Equal(t, toolB, myStats[1].Name)
	assert.Equal(t, int64(10), myStats[1].TotalCalls)
	assert.Equal(t, int64(10), myStats[1].SuccessCount)
	assert.Equal(t, int64(0), myStats[1].ErrorCount)
	assert.Equal(t, 0.0, myStats[1].FailureRate)
}
