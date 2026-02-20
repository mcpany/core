// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/topology"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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
		MetricsGatherer: registry, statsCache: make(map[string]statsCacheEntry),
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

func TestHandleDashboardTopTools(t *testing.T) {
	// Define counters
	toolsCallTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcpany_tools_call_total",
			Help: "Total number of tool calls",
		},
		[]string{"tool", "service_id", "status", "error_type"},
	)

	registry := prometheus.NewRegistry()
	require.NoError(t, registry.Register(toolsCallTotal))

	tool1 := "top_tool_1"
	tool2 := "top_tool_2"

	// Tool 1: 100 calls
	toolsCallTotal.WithLabelValues(tool1, "service1", "success", "").Add(100)

	// Tool 2: 50 calls
	toolsCallTotal.WithLabelValues(tool2, "service1", "success", "").Add(50)

	app := &Application{
		MetricsGatherer: registry, statsCache: make(map[string]statsCacheEntry),
	}

	handler := app.handleDashboardTopTools()
	req, _ := http.NewRequest("GET", "/dashboard/top-tools", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var stats []ToolUsageStats
	err := json.Unmarshal(rr.Body.Bytes(), &stats)
	require.NoError(t, err)

	// Find our tools
	var t1, t2 *ToolUsageStats
	for i := range stats {
		if stats[i].Name == tool1 {
			t1 = &stats[i]
		} else if stats[i].Name == tool2 {
			t2 = &stats[i]
		}
	}

	require.NotNil(t, t1)
	require.NotNil(t, t2)
	assert.Equal(t, int64(100), t1.Count)
	assert.Equal(t, int64(50), t2.Count)
}

func TestHandleDashboardTraffic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRegistry := new(MockServiceRegistry)     // from api_error_test.go
	mockTM := tool.NewMockManagerInterface(ctrl) // from server/pkg/tool

	// We need real Topology Manager logic, so we use real NewManager
	tm := topology.NewManager(mockRegistry, mockTM)

	app := &Application{
		TopologyManager: tm, statsCache: make(map[string]statsCacheEntry),
	}

	// Seed traffic relative to now to ensure it appears in GetTrafficHistory
	nowStr := time.Now().Add(-5 * time.Minute).Format("15:04")
	tm.SeedTrafficHistory([]topology.TrafficPoint{
		{Time: nowStr, Total: 100, Latency: 50},
	})

	req, _ := http.NewRequest("GET", "/dashboard/traffic", nil)
	rr := httptest.NewRecorder()

	handler := app.handleDashboardTraffic()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var points []topology.TrafficPoint
	err := json.Unmarshal(rr.Body.Bytes(), &points)
	require.NoError(t, err)

	assert.NotEmpty(t, points)
	// Check if we find the seeded point (total 100)
	found := false
	for _, p := range points {
		if p.Total == 100 {
			found = true
			break
		}
	}
	assert.True(t, found, "Should find the seeded point")
}

func TestHandleDebugSeedTraffic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRegistry := new(MockServiceRegistry)
	mockTM := tool.NewMockManagerInterface(ctrl)
	tm := topology.NewManager(mockRegistry, mockTM)

	app := &Application{
		TopologyManager: tm, statsCache: make(map[string]statsCacheEntry),
	}

	points := []topology.TrafficPoint{
		{Time: "10:00", Total: 123, Latency: 10},
	}
	body, _ := json.Marshal(points)

	req, _ := http.NewRequest("POST", "/debug/seed_traffic", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler := app.handleDebugSeedTraffic()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	stats := tm.GetStats("")
	assert.Equal(t, int64(123), stats.TotalRequests)
}

type TestMockPrompt struct {
	name string
}

func (m *TestMockPrompt) Prompt() *mcp.Prompt {
	return &mcp.Prompt{Name: m.name}
}
func (m *TestMockPrompt) Service() string { return "test" }
func (m *TestMockPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	return nil, nil
}

type TestMockResource struct {
	uri string
}

func (m *TestMockResource) Resource() *mcp.Resource { return &mcp.Resource{URI: m.uri} }
func (m *TestMockResource) Service() string         { return "test" }
func (m *TestMockResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	return nil, nil
}
func (m *TestMockResource) Subscribe(ctx context.Context) error { return nil }

func TestHandleDashboardMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRegistry := new(MockServiceRegistry)
	mockTM := tool.NewMockManagerInterface(ctrl)

	// Mock Managers for Counts
	mockRegistry.On("GetAllServices").Return(func() []*configv1.UpstreamServiceConfig {
		s := &configv1.UpstreamServiceConfig{}
		s.SetName("s1")
		return []*configv1.UpstreamServiceConfig{s}
	}(), nil)
	mockTM.EXPECT().ListTools().Return([]tool.Tool{&TestMockTool{}})

	// Topology
	tm := topology.NewManager(mockRegistry, mockTM)
	tm.SeedTrafficHistory([]topology.TrafficPoint{
		{Time: "12:00", Total: 60, Latency: 100},
	})

	mockPM := prompt.NewMockManagerInterface(ctrl)
	mockRM := resource.NewMockManagerInterface(ctrl)

	mockPM.EXPECT().ListPrompts().Return([]prompt.Prompt{&TestMockPrompt{name: "p1"}})
	mockRM.EXPECT().ListResources().Return([]resource.Resource{&TestMockResource{uri: "r1"}})

	app := &Application{
		TopologyManager: tm, statsCache: make(map[string]statsCacheEntry),
		ServiceRegistry: mockRegistry,
		ToolManager:     mockTM,
		PromptManager:   mockPM,
		ResourceManager: mockRM,
	}

	handler := app.handleDashboardMetrics()

	req, _ := http.NewRequest("GET", "/dashboard/metrics", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var metrics []Metric
	err := json.Unmarshal(rr.Body.Bytes(), &metrics)
	require.NoError(t, err)

	// Verify values
	metricMap := make(map[string]string)
	for _, m := range metrics {
		metricMap[m.Label] = m.Value
	}

	assert.Equal(t, "1", metricMap["Active Services"])
	assert.Equal(t, "1", metricMap["Connected Tools"])
	assert.Equal(t, "1", metricMap["Prompts"])
	assert.Equal(t, "1", metricMap["Resources"])
	assert.Equal(t, "60", metricMap["Total Requests"])
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

	toolA := "tool_usage_A"
	toolB := "tool_usage_B"

	// Tool A: 5 success, 5 error => 50% success
	toolsCallTotal.WithLabelValues(toolA, "service1", "success", "").Add(5)
	toolsCallTotal.WithLabelValues(toolA, "service1", "error", "some_error").Add(5)

	// Tool B: 10 success, 0 error => 100% success
	toolsCallTotal.WithLabelValues(toolB, "service1", "success", "").Add(10)

	// Create Request
	req, err := http.NewRequest("GET", "/dashboard/tool-usage", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	app := &Application{
		MetricsGatherer: registry, statsCache: make(map[string]statsCacheEntry),
	}

	handler := app.handleDashboardToolUsage()
	handler.ServeHTTP(rr, req)

	// Check Response
	assert.Equal(t, http.StatusOK, rr.Code)

	var stats []ToolAnalytics
	err = json.Unmarshal(rr.Body.Bytes(), &stats)
	require.NoError(t, err)

	// Filter stats
	var myStats []ToolAnalytics
	for _, s := range stats {
		if s.Name == toolA || s.Name == toolB {
			myStats = append(myStats, s)
		}
	}

	// Sort by Name (A then B)
	require.Equal(t, 2, len(myStats))
	assert.Equal(t, toolA, myStats[0].Name)
	assert.Equal(t, int64(10), myStats[0].TotalCalls)
	assert.Equal(t, 50.0, myStats[0].SuccessRate)

	assert.Equal(t, toolB, myStats[1].Name)
	assert.Equal(t, int64(10), myStats[1].TotalCalls)
	assert.Equal(t, 100.0, myStats[1].SuccessRate)
}

func TestStatsCacheEviction(t *testing.T) {
	app := &Application{
		statsCache: make(map[string]statsCacheEntry),
	}

	// Fill cache to limit (100)
	for i := 0; i < 100; i++ {
		app.setStatsCache(fmt.Sprintf("key-%d", i), i)
	}

	// Assert full
	assert.Equal(t, 100, len(app.statsCache))

	// Add 101st item, triggering eviction
	app.setStatsCache("key-101", 101)

	// Previously it would be 1. Now it should be ~75 + 1 = 76.
	// We allow some flexibility because implementation details might change slightly,
	// but it should definitely be > 50 and < 90.
	size := len(app.statsCache)
	assert.Greater(t, size, 50, "Cache should not be cleared completely")
	assert.Less(t, size, 90, "Cache should have evicted some items")

	// Also ensure key-101 is present
	val, ok := app.getStatsCache("key-101")
	assert.True(t, ok)
	assert.Equal(t, 101, val)
}

func TestCalculateUptime(t *testing.T) {
	window := 24 * time.Hour
	now := time.Now().UnixMilli()

	tests := []struct {
		name     string
		history  []health.HistoryPoint
		expected string
	}{
		{
			name:     "Empty history",
			history:  []health.HistoryPoint{},
			expected: "100.0%",
		},
		{
			name: "Full uptime (1 point long ago)",
			history: []health.HistoryPoint{
				{Timestamp: now - window.Milliseconds() - 1000, Status: "UP"},
			},
			expected: "100.0%",
		},
		{
			name: "Full downtime (1 point long ago)",
			history: []health.HistoryPoint{
				{Timestamp: now - window.Milliseconds() - 1000, Status: "DOWN"},
			},
			expected: "0.0%",
		},
		{
			name: "50% uptime (UP -> DOWN halfway)",
			history: []health.HistoryPoint{
				{Timestamp: now - window.Milliseconds(), Status: "UP"},
				{Timestamp: now - (window.Milliseconds() / 2), Status: "DOWN"},
			},
			expected: "50.0%",
		},
		{
			name: "Mixed states",
			history: []health.HistoryPoint{
				// Start DOWN
				{Timestamp: now - window.Milliseconds(), Status: "DOWN"},
				// UP after 25% of window
				{Timestamp: now - (window.Milliseconds() * 3 / 4), Status: "UP"},
				// DOWN after 75% of window (so UP for 50%)
				{Timestamp: now - (window.Milliseconds() / 4), Status: "DOWN"},
			},
			// UP duration: from 3/4 to 1/4 = 50%
			expected: "50.0%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mocking time.Now() is hard without dependency injection or clock interface.
			// However, our helper uses time.Now() inside.
			// We can bypass this by ensuring our test timestamps are relative to ACTUAL time.Now()
			// But calculateUptime calls time.Now() internally.
			// Ideally we refactor calculateUptime to accept `now` or use a clock.
			// For this test, since we use `time.Now()` to generate test data, it should align roughly.
			// But there is a small race condition if seconds tick over.
			// Given it uses Milliseconds, it might be flaky if execution takes significant time.
			// But logical time is derived from `now` variable which is computed at test start.
			// `calculateUptime` will compute a NEW `now`.
			// The gap between test setup and execution might introduce error.
			// To be robust, I should pass `now` to `calculateUptime` or inject it.
			// But for now, let's assume the delta is negligible for the math (0 vs small ms).
			// ACTUALLY, strict equality "50.0%" might fail if 1ms off.
			// Let's refactor calculateUptime to accept `now`? No, that changes signature for production code.
			// Let's just make the window large enough that few ms doesn't change percentage?
			// Window is 24h. 100ms is negligible.

			// However, in `calculateUptime`: `now := time.Now().UnixMilli()`
			// If I construct points relative to `now` in test, but `calculateUptime` runs 10ms later,
			// the window shifts by 10ms.
			// If my points are exactly on the boundary, it matters.
			// My test points are "now - window".
			// If `calculateUptime`'s now is `now + 10ms`, then `startTime` is `now + 10ms - window`.
			// My point `now - window` is definitely < `startTime`.
			// So it picks up the initial status. Correct.

			// But the *end* point is `now + 10ms`.
			// My last point might be `now - window/4`.
			// The duration calculation uses `calculateUptime`'s now.
			// So total window is 24h.
			// The error should be very small.

			result := calculateUptime(tt.history, window)
			assert.Equal(t, tt.expected, result)
		})
	}
}
