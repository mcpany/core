// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/topology"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// MockServiceRegistryForDashboard is a mock implementation of ServiceRegistryInterface.
type MockServiceRegistryForDashboard struct {
	mock.Mock
}

func (m *MockServiceRegistryForDashboard) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	args := m.Called(ctx, serviceConfig)
	return args.String(0), nil, nil, args.Error(3)
}

func (m *MockServiceRegistryForDashboard) UnregisterService(ctx context.Context, serviceName string) error {
	args := m.Called(ctx, serviceName)
	return args.Error(0)
}

func (m *MockServiceRegistryForDashboard) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	args := m.Called()
	return args.Get(0).([]*configv1.UpstreamServiceConfig), args.Error(1)
}

func (m *MockServiceRegistryForDashboard) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	return nil, args.Bool(1)
}

func (m *MockServiceRegistryForDashboard) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	args := m.Called(serviceID)
	return nil, args.Bool(1)
}

func (m *MockServiceRegistryForDashboard) GetServiceError(serviceID string) (string, bool) {
	args := m.Called(serviceID)
	return args.String(0), args.Bool(1)
}

func TestHandleDashboardHealth_Integration(t *testing.T) {
	// Setup TopologyManager
	tm := topology.NewManager(nil, nil)
	defer tm.Close()

	// Seed TopologyManager
	serviceID := "test-service-id"

	// Record some activity (200ms latency)
	tm.RecordActivity("session-1", nil, 200*time.Millisecond, false, serviceID)

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	mockRegistry := new(MockServiceRegistryForDashboard)

	svc := configv1.UpstreamServiceConfig_builder{
		Id:   proto.String(serviceID),
		Name: proto.String("test-service"),
	}.Build()

	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{svc}, nil)
	mockRegistry.On("GetServiceError", serviceID).Return("", false)

	app := &Application{
		ServiceRegistry: mockRegistry,
		TopologyManager: tm,
	}

	req := httptest.NewRequest(http.MethodGet, "/dashboard/health", nil)
	w := httptest.NewRecorder()

	handler := app.handleDashboardHealth()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp ServiceHealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Len(t, resp.Services, 1)
	svcResp := resp.Services[0]
	assert.Equal(t, serviceID, svcResp.ID)
	// Latency should be "200ms"
	assert.Equal(t, "200ms", svcResp.Latency)

	// Uptime will likely be "0.0%" or "Unknown" because history is empty.
	assert.Equal(t, "Unknown", svcResp.Uptime)
}
