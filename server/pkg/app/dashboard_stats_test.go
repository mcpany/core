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
	now := time.Now().UnixMilli()
	window := 24 * time.Hour

	tests := []struct {
		name     string
		history  []health.HistoryPoint
		expected string
	}{
		{
			name:     "No history",
			history:  []health.HistoryPoint{},
			expected: "N/A",
		},
		{
			name: "Always UP (starts before window)",
			history: []health.HistoryPoint{
				{Timestamp: now - window.Milliseconds() - 1000, Status: "up"},
			},
			expected: "100%",
		},
		{
			name: "Always UP (starts within window)",
			history: []health.HistoryPoint{
				{Timestamp: now - (12 * time.Hour).Milliseconds(), Status: "up"},
			},
			// From start to 12h ago: unknown. From 12h ago to now: up.
			// 12h up / 24h = 50%
			expected: "50.0%",
		},
		{
			name: "Half UP (starts before window, goes down halfway)",
			history: []health.HistoryPoint{
				{Timestamp: now - window.Milliseconds() - 1000, Status: "up"},
				{Timestamp: now - (12 * time.Hour).Milliseconds(), Status: "down"},
			},
			// From start to 12h ago: up (12h). From 12h ago to now: down (12h).
			expected: "50.0%",
		},
		{
			name: "Mixed states",
			history: []health.HistoryPoint{
				{Timestamp: now - (20 * time.Hour).Milliseconds(), Status: "up"},   // Up for 10h
				{Timestamp: now - (10 * time.Hour).Milliseconds(), Status: "down"}, // Down for 5h
				{Timestamp: now - (5 * time.Hour).Milliseconds(), Status: "up"},    // Up for 5h
			},
			// Window: 24h.
			// 0h-4h: Unknown (starts at 20h ago)
			// 4h-14h (10h): UP
			// 14h-19h (5h): DOWN
			// 19h-24h (5h): UP
			// Total UP: 15h. Total: 24h.
			// 15/24 = 62.5%
			expected: "62.5%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateUptime(tt.history, window)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandleDashboardHealth_RealStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRegistry := new(MockServiceRegistry)
	mockTM := tool.NewMockManagerInterface(ctrl)

	// Real Topology Manager
	tm := topology.NewManager(mockRegistry, mockTM)

	// Seed stats for service "s1"
	tm.RecordActivity("sess1", nil, 123*time.Millisecond, false, "s1")
	// Wait for async processing
	assert.Eventually(t, func() bool {
		lat, _ := tm.GetRecentServiceStats("s1", 1*time.Minute)
		return lat > 0
	}, 1*time.Second, 10*time.Millisecond)

	mockRegistry.On("GetAllServices").Return(func() []*configv1.UpstreamServiceConfig {
		s := &configv1.UpstreamServiceConfig{}
		s.SetName("s1")
		s.SetId("s1")
		return []*configv1.UpstreamServiceConfig{s}
	}(), nil)
	mockRegistry.On("GetServiceError", "s1").Return("", false)

	// Inject Health History
	health.AddHealthStatus("s1", "up")

	app := &Application{
		TopologyManager: tm,
		ServiceRegistry: mockRegistry,
	}

	handler := app.handleDashboardHealth()
	req, _ := http.NewRequest("GET", "/dashboard/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp ServiceHealthResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.Len(t, resp.Services, 1)
	svc := resp.Services[0]
	assert.Equal(t, "s1", svc.Name)
	assert.Equal(t, "123ms", svc.Latency)
	assert.Contains(t, []string{"0.0%", "100%"}, svc.Uptime)
}
