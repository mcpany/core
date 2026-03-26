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
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/topology"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func TestHandleDashboardMetrics_EdgeCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("Nil TopologyManager", func(t *testing.T) {
		app := &Application{
			TopologyManager: nil, // Explicitly nil
		}

		req := httptest.NewRequest(http.MethodGet, "/dashboard/metrics", nil)
		w := httptest.NewRecorder()

		handler := app.handleDashboardMetrics()
		handler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var metrics []Metric
		err := json.NewDecoder(resp.Body).Decode(&metrics)
		require.NoError(t, err)

		metricsMap := make(map[string]Metric)
		for _, m := range metrics {
			metricsMap[m.Label] = m
		}

		// Should default to 0/empty
		assert.Equal(t, "0", metricsMap["Total Requests"].Value)
		assert.Equal(t, "0ms", metricsMap["Avg Latency"].Value)
		assert.Equal(t, "0.00%", metricsMap["Error Rate"].Value)
	})

	t.Run("ServiceRegistry Error", func(t *testing.T) {
		mockRegistry := new(MockServiceRegistry) // Reusing MockServiceRegistry from api_test.go
		mockTM := tool.NewMockManagerInterface(ctrl)

		app := &Application{
			ServiceRegistry: mockRegistry,
			ToolManager:     mockTM,
		}

		// Simulate error from GetAllServices
		// Note: We must cast nil to the slice type to avoid panic in the mock's type assertion
		mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig(nil), assert.AnError)

		// ToolManager returns nil tools
		mockTM.EXPECT().ListTools().Return(nil)

		req := httptest.NewRequest(http.MethodGet, "/dashboard/metrics", nil)
		w := httptest.NewRecorder()

		handler := app.handleDashboardMetrics()
		handler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var metrics []Metric
		err := json.NewDecoder(resp.Body).Decode(&metrics)
		require.NoError(t, err)

		metricsMap := make(map[string]Metric)
		for _, m := range metrics {
			metricsMap[m.Label] = m
		}

		// Service count should be 0 because error was ignored
		assert.Equal(t, "0", metricsMap["Active Services"].Value)
	})

	t.Run("Managers Nil", func(t *testing.T) {
		// All managers nil
		app := &Application{}

		req := httptest.NewRequest(http.MethodGet, "/dashboard/metrics", nil)
		w := httptest.NewRecorder()

		handler := app.handleDashboardMetrics()
		handler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var metrics []Metric
		err := json.NewDecoder(resp.Body).Decode(&metrics)
		require.NoError(t, err)

		metricsMap := make(map[string]Metric)
		for _, m := range metrics {
			metricsMap[m.Label] = m
		}

		assert.Equal(t, "0", metricsMap["Active Services"].Value)
		assert.Equal(t, "0", metricsMap["Connected Tools"].Value)
		assert.Equal(t, "0", metricsMap["Resources"].Value)
		assert.Equal(t, "0", metricsMap["Prompts"].Value)
	})

	t.Run("Topology with Traffic History", func(t *testing.T) {
		// This verifies throughput calculation
		mockRegistry := new(MockServiceRegistry)
		mockTM := tool.NewMockManagerInterface(ctrl)
		mockPM := prompt.NewMockManagerInterface(ctrl)
		mockRM := resource.NewMockManagerInterface(ctrl)

		topoManager := topology.NewManager(mockRegistry, mockTM)
		defer topoManager.Close()

		// Seed 120 requests in ONE minute (e.g. now).
		// GetTrafficHistory returns 60 points (last 60 mins).
		// One point has 120 requests. Others 0.
		// Total in window = 120.
		// Window duration = 60 minutes = 3600 seconds.
		// Throughput = 120 / 3600 = 0.0333 rps.

		topoManager.SeedTrafficHistory([]topology.TrafficPoint{
			{Time: time.Now().Format("15:04"), Total: 120, Latency: 10},
		})

		// Wait for sync
		require.Eventually(t, func() bool {
			return topoManager.GetStats("").TotalRequests == 120
		}, 1*time.Second, 10*time.Millisecond)

		app := &Application{
			TopologyManager: topoManager,
			ServiceRegistry: mockRegistry,
			ToolManager:     mockTM,
			PromptManager:   mockPM,
			ResourceManager: mockRM,
		}

		mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig(nil), nil)
		mockTM.EXPECT().ListTools().Return(nil)
		mockPM.EXPECT().ListPrompts().Return(nil)
		mockRM.EXPECT().ListResources().Return(nil)

		req := httptest.NewRequest(http.MethodGet, "/dashboard/metrics", nil)
		w := httptest.NewRecorder()

		handler := app.handleDashboardMetrics()
		handler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		var metrics []Metric
		err := json.NewDecoder(resp.Body).Decode(&metrics)
		require.NoError(t, err)

		metricsMap := make(map[string]Metric)
		for _, m := range metrics {
			metricsMap[m.Label] = m
		}

		// 120 / 3600 = 0.0333
		assert.Contains(t, metricsMap["Avg Throughput"].Value, "0.03 rps")
	})
}
