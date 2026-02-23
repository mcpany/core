// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/topology"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestDashboardMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mocks
	mockRegistry := new(serviceregistry.MockServiceRegistry)
	mockToolManager := tool.NewMockManagerInterface(ctrl)
	mockResourceManager := resource.NewMockManagerInterface(ctrl)
	mockPromptManager := prompt.NewMockManagerInterface(ctrl)

	// Setup Topology Manager
	// We need to allow calls to GetAllServices and ListTools for GetGraph if it's called internally,
	// but handleDashboardMetrics calls GetStats and GetTrafficHistory which use internal state.
	// However, handleDashboardMetrics ALSO calls ListTools, ListResources, ListPrompts directly on managers.

	// Setup expectations for handleDashboardMetrics
	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{
		configv1.UpstreamServiceConfig_builder{Name: proto.String("service-1")}.Build(),
		configv1.UpstreamServiceConfig_builder{Name: proto.String("service-2")}.Build(),
	}, nil)

	// Dashboard only counts the items, so returning empty lists is sufficient for verifying the flow.
	mockToolManager.EXPECT().ListTools().Return([]tool.Tool{}).Times(1)
	mockResourceManager.EXPECT().ListResources().Return([]resource.Resource{}).Times(1)
	mockPromptManager.EXPECT().ListPrompts().Return([]prompt.Prompt{}).Times(1)

	// Create Topology Manager and Seed Data
	// Note: TopologyManager.NewManager spawns a goroutine. We should close it.
	tm := topology.NewManager(mockRegistry, mockToolManager)
	defer tm.Close()

	// Seed Traffic History
	// 100 requests total, 10 errors, 50ms average latency
	// We use current time to ensure it falls within the last 60 minutes window
	now := time.Now()
	timeStr := now.Format("15:04")
	points := []topology.TrafficPoint{
		{Time: timeStr, Total: 100, Errors: 10, Latency: 50},
	}
	tm.SeedTrafficHistory(points)

	// Construct Application
	app := &Application{
		TopologyManager: tm,
		ServiceRegistry: mockRegistry,
		ToolManager:     mockToolManager,
		ResourceManager: mockResourceManager,
		PromptManager:   mockPromptManager,
	}

	// Create Request
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/metrics", nil)
	w := httptest.NewRecorder()

	// Execute
	handler := app.handleDashboardMetrics()
	handler.ServeHTTP(w, req)

	// Verify Response
	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var metrics []Metric
	err := json.NewDecoder(resp.Body).Decode(&metrics)
	require.NoError(t, err)

	// Verify Metrics Content
	// We expect 8 metrics
	assert.Len(t, metrics, 8)

	metricsMap := make(map[string]Metric)
	for _, m := range metrics {
		metricsMap[m.Label] = m
	}

	// 1. Total Requests
	// Seeded: 100 requests
	assert.Equal(t, "100", metricsMap["Total Requests"].Value)

	// 2. Avg Throughput
	// Seeded: 100 requests in one minute bucket within the last 60 minutes.
	// GetTrafficHistory returns 60 points (60 minutes).
	// Throughput = Total Requests / (60 minutes * 60 seconds) = 100 / 3600 = 0.0277... -> 0.03 rps
	assert.Equal(t, "0.03 rps", metricsMap["Avg Throughput"].Value)

	// 3. Active Services
	// Mock returned 2 services
	assert.Equal(t, "2", metricsMap["Active Services"].Value)

	// 4. Connected Tools
	// Mock returned 0
	assert.Equal(t, "0", metricsMap["Connected Tools"].Value)

	// 5. Resources
	// Mock returned 0
	assert.Equal(t, "0", metricsMap["Resources"].Value)

	// 6. Prompts
	// Mock returned 0
	assert.Equal(t, "0", metricsMap["Prompts"].Value)

	// 7. Avg Latency
	// Seeded: 50ms (average latency in SeedTrafficHistory logic)
	// Wait, SeedTrafficHistory implementation:
	// m.trafficHistory[targetTime.Unix()] = &MinuteStats{
	// 	Requests: p.Total,
	// 	Errors:   p.Errors,
	// 	Latency:  p.Latency * p.Total, // Reverse average
	// }
	// And GetStats implementation:
	// avgLatency = time.Duration(int64(totalLatency) / totalRequests)
	// So (50 * 100) / 100 = 50.
	assert.Equal(t, "50ms", metricsMap["Avg Latency"].Value)

	// 8. Error Rate
	// Seeded: 10 errors / 100 requests = 10%
	assert.Equal(t, "10.00%", metricsMap["Error Rate"].Value)
}

func TestDashboardMetrics_MethodNotAllowed(t *testing.T) {
	app := &Application{}
	req := httptest.NewRequest(http.MethodPost, "/api/dashboard/metrics", nil)
	w := httptest.NewRecorder()

	handler := app.handleDashboardMetrics()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
}
