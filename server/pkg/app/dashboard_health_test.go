// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleDashboardHealth(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)

    // Setup services with unique names to avoid collision in global health history
    svcHealthyName := "svc_healthy_" + t.Name()
    svcUnhealthyName := "svc_unhealthy_" + t.Name()
    svcInactiveName := "svc_inactive_" + t.Name()
    svcUnknownName := "svc_unknown_" + t.Name()
    svcDegradedName := "svc_degraded_" + t.Name()

    // Cleanup function
    t.Cleanup(func() {
        health.RemoveHealthStatus(svcHealthyName)
        health.RemoveHealthStatus(svcUnhealthyName)
        health.RemoveHealthStatus(svcInactiveName)
        health.RemoveHealthStatus(svcUnknownName)
        health.RemoveHealthStatus(svcDegradedName)
    })

	// Seed Health History
    health.AddHealthStatus(svcHealthyName, "UP")
    health.AddHealthStatus(svcUnhealthyName, "DOWN")
    // svcInactive, svcUnknown, svcDegraded have no history or specific history handling
    health.AddHealthStatus(svcDegradedName, "UP")

	// Define Services
	svcHealthy := &configv1.UpstreamServiceConfig{}
	svcHealthy.SetId("id_healthy")
	svcHealthy.SetName(svcHealthyName)

	svcUnhealthy := &configv1.UpstreamServiceConfig{}
	svcUnhealthy.SetId("id_unhealthy")
	svcUnhealthy.SetName(svcUnhealthyName)

	svcInactive := &configv1.UpstreamServiceConfig{}
	svcInactive.SetId("id_inactive")
	svcInactive.SetName(svcInactiveName)
	svcInactive.SetDisable(true)

	svcUnknown := &configv1.UpstreamServiceConfig{}
	svcUnknown.SetId("id_unknown")
	svcUnknown.SetName(svcUnknownName)

	svcDegraded := &configv1.UpstreamServiceConfig{}
	svcDegraded.SetId("id_degraded")
	svcDegraded.SetName(svcDegradedName)

	services := []*configv1.UpstreamServiceConfig{
		svcHealthy, svcUnhealthy, svcInactive, svcUnknown, svcDegraded,
	}

	// Mock Registry Behavior
	mockRegistry.On("GetAllServices").Return(services, nil)

    // Mock GetServiceError
    mockRegistry.On("GetServiceError", "id_healthy").Return("", false)
    mockRegistry.On("GetServiceError", "id_unhealthy").Return("", false)
    mockRegistry.On("GetServiceError", "id_inactive").Return("", false)
    mockRegistry.On("GetServiceError", "id_unknown").Return("", false)
    mockRegistry.On("GetServiceError", "id_degraded").Return("partial failure", true)

	app := &Application{
		ServiceRegistry: mockRegistry,
	}

	req, err := http.NewRequest("GET", "/dashboard/health", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := app.handleDashboardHealth()
	handler.ServeHTTP(rr, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp ServiceHealthResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

    // Convert slice to map for easy lookup
    svcMap := make(map[string]ServiceHealth)
    for _, s := range resp.Services {
        svcMap[s.Name] = s
    }

    // 1. Healthy Service
    s, ok := svcMap[svcHealthyName]
    require.True(t, ok)
    assert.Equal(t, "healthy", s.Status)
    assert.Equal(t, "id_healthy", s.ID)

    // 2. Unhealthy Service
    s, ok = svcMap[svcUnhealthyName]
    require.True(t, ok)
    assert.Equal(t, "unhealthy", s.Status)

    // 3. Inactive Service
    s, ok = svcMap[svcInactiveName]
    require.True(t, ok)
    assert.Equal(t, "inactive", s.Status)

    // 4. Unknown Service (no history, enabled, no error)
    s, ok = svcMap[svcUnknownName]
    require.True(t, ok)
    assert.Equal(t, "unknown", s.Status)

    // 5. Degraded Service (UP but has error)
    s, ok = svcMap[svcDegradedName]
    require.True(t, ok)
    assert.Equal(t, "degraded", s.Status)
    assert.Equal(t, "partial failure", s.Message)

    // Check History Presence
    assert.NotEmpty(t, resp.History["id_healthy"])
    assert.NotEmpty(t, resp.History["id_unhealthy"])
}
