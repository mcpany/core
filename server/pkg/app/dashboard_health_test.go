// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/topology"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHandleDashboardHealth_RealStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRegistry := new(MockServiceRegistry)
	mockTM := tool.NewMockManagerInterface(ctrl)

	// Mock Service Registry to return one service
	s := &configv1.UpstreamServiceConfig{}
	s.SetName("my-service")
	s.SetId("my-service-id")

	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{s}, nil)
	mockRegistry.On("GetServiceError", "my-service-id").Return("", false)

	// Real Topology Manager
	tm := topology.NewManager(mockRegistry, mockTM)
	defer tm.Close()
	defer health.ResetHealthHistory()

	// Seed Traffic for Latency using RecordActivity
	// We use "my-service" as serviceID because handleDashboardHealth uses svc.GetName()
	// Record 10 requests with 123ms latency
	for i := 0; i < 10; i++ {
		tm.RecordActivity("session-1", nil, 123*time.Millisecond, false, "my-service")
	}

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	// Seed Health History for Uptime
	// Simulating: "up" for last 25h.
	now := time.Now()
	start := now.Add(-25 * time.Hour).UnixMilli()

	health.SeedHealthHistory("my-service", []health.HistoryPoint{
		{Timestamp: start, Status: "up"},
	})

	app := &Application{
		TopologyManager: tm,
		ServiceRegistry: mockRegistry,
		statsCache:      make(map[string]statsCacheEntry),
	}

	// Execute Request
	req, _ := http.NewRequest("GET", "/dashboard/health", nil)
	rr := httptest.NewRecorder()

	handler := app.handleDashboardHealth()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp ServiceHealthResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.Len(t, resp.Services, 1)
	svc := resp.Services[0]

	assert.Equal(t, "my-service", svc.Name)

	// Check Latency
	// 123ms seeded. Should return "123ms".
	assert.Equal(t, "123ms", svc.Latency)

	// Check Uptime
	// Seeded "up" at -25h. Window is 24h. Should be 100.0%.
	assert.Equal(t, "100.0%", svc.Uptime)
}
