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

func TestHandleDashboardHealth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// MockServiceRegistry should be available in the package test scope
	// If not, we might need to define a simple mock here.
	// Assuming it's available as per previous tests (e.g. TestHandleDashboardTraffic)
	mockRegistry := new(MockServiceRegistry)

	mockTM := tool.NewMockManagerInterface(ctrl)

	// Mock Service Registry behavior
	mockRegistry.On("GetAllServices").Return(func() []*configv1.UpstreamServiceConfig {
		s := &configv1.UpstreamServiceConfig{}
		s.SetName("service-1")
		s.SetId("service-1")
		return []*configv1.UpstreamServiceConfig{s}
	}(), nil)

	mockRegistry.On("GetServiceError", "service-1").Return("", false)

	// Setup Topology Manager with seeded data
	// Note: We use the real topology manager with mocks injected
	tm := topology.NewManager(mockRegistry, mockTM)

	// Seed traffic history
	now := time.Now()
	nowStr := now.Format("15:04")
	tm.SeedTrafficHistory([]topology.TrafficPoint{
		{Time: nowStr, Total: 10, Latency: 100, ServiceID: "service-1"},
	})

	// Setup Health History
	// health.AddHealthStatus populates global state.
	// Ideally we reset it, but we can just use a unique service name if we were worried about pollution.
	// But let's use "service-1".
	health.AddHealthStatusWithTime("service-1", "UP", now.Add(-24*time.Hour))

	app := &Application{
		ServiceRegistry: mockRegistry,
		TopologyManager: tm,
		statsCache:      make(map[string]statsCacheEntry),
	}

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
	assert.Equal(t, "service-1", svc.Name)
	assert.Equal(t, "healthy", svc.Status)

	// Verify Latency
	// 100ms should be formatted as "100ms"
	assert.Equal(t, "100ms", svc.Latency)

	// Verify Uptime
	// With one "UP" point, it should be 100%
	assert.Equal(t, "100.0%", svc.Uptime)
}
