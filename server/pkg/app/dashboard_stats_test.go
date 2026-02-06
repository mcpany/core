package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
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
