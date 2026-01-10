// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// MockServiceRegistry is a mock implementation of ServiceRegistryInterface
type MockServiceRegistry struct {
	mock.Mock
}

func (m *MockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	args := m.Called(ctx, serviceConfig)
	return args.String(0), args.Get(1).([]*configv1.ToolDefinition), args.Get(2).([]*configv1.ResourceDefinition), args.Error(3)
}

func (m *MockServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	args := m.Called(ctx, serviceName)
	return args.Error(0)
}

func (m *MockServiceRegistry) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	args := m.Called()
	return args.Get(0).([]*configv1.UpstreamServiceConfig), args.Error(1)
}

func (m *MockServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	if info := args.Get(0); info != nil {
		return info.(*tool.ServiceInfo), args.Bool(1)
	}
	return nil, args.Bool(1)
}

func (m *MockServiceRegistry) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	args := m.Called(serviceID)
	if cfg := args.Get(0); cfg != nil {
		return cfg.(*configv1.UpstreamServiceConfig), args.Bool(1)
	}
	return nil, args.Bool(1)
}

func (m *MockServiceRegistry) GetServiceError(serviceID string) (string, bool) {
	args := m.Called(serviceID)
	return args.String(0), args.Bool(1)
}

func TestHandleServices_IncludesError(t *testing.T) {
	// Setup DB
	db, err := sqlite.NewDB(":memory:")
	require.NoError(t, err)
	defer db.Close()
	store := sqlite.NewStore(db)

	// Setup Mock Registry
	mockRegistry := new(MockServiceRegistry)
	service1 := &configv1.UpstreamServiceConfig{
		Name: proto.String("service-1"),
		Id:   proto.String("service-1"),
	}
	service2 := &configv1.UpstreamServiceConfig{
		Name: proto.String("service-2"),
		Id:   proto.String("service-2"),
	}
	// Service 3 has no ID but has sanitized name
	service3 := &configv1.UpstreamServiceConfig{
		Name:          proto.String("service-3"),
		SanitizedName: proto.String("service-3-sanitized"),
	}

	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{service1, service2, service3}, nil)
	mockRegistry.On("GetServiceError", "service-1").Return("", false)
	mockRegistry.On("GetServiceError", "service-2").Return("Connection refused", true)
	mockRegistry.On("GetServiceError", "service-3-sanitized").Return("Another error", true)

	// Setup Application
	app := NewApplication()
	app.ServiceRegistry = mockRegistry

	handler := app.handleServices(store)
	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var services []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&services)
	require.NoError(t, err)

	assert.Len(t, services, 3)

	var s1, s2, s3 map[string]interface{}
	for _, s := range services {
		if s["name"] == "service-1" {
			s1 = s
		} else if s["name"] == "service-2" {
			s2 = s
		} else if s["name"] == "service-3" {
			s3 = s
		}
	}

	assert.NotNil(t, s1)
	assert.NotNil(t, s2)
	assert.NotNil(t, s3)

	// Check errors
	assert.Nil(t, s1["last_error"], "service-1 should not have error")
	assert.Equal(t, "Connection refused", s2["last_error"], "service-2 should have error")
	assert.Equal(t, "Another error", s3["last_error"], "service-3 should have error")
}
