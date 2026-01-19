// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

type mockServiceRegistry struct {
	services []*configv1.UpstreamServiceConfig
	err      error
}

func (m *mockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	return "", nil, nil, nil
}
func (m *mockServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	return nil
}
func (m *mockServiceRegistry) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	return m.services, m.err
}
func (m *mockServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return nil, false
}
func (m *mockServiceRegistry) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	return nil, false
}
func (m *mockServiceRegistry) GetServiceError(serviceID string) (string, bool) {
	return "", false
}

func TestServicesHealthCheck(t *testing.T) {
	app := NewApplication()

	// Test 1: Registry not initialized
	res := app.servicesHealthCheck(context.Background())
	assert.Equal(t, "ok", res.Status)
	assert.Contains(t, res.Message, "not initialized")

	// Test 2: Error getting services
	mockReg := &mockServiceRegistry{
		err: errors.New("db error"),
	}
	app.ServiceRegistry = mockReg
	res = app.servicesHealthCheck(context.Background())
	assert.Equal(t, "degraded", res.Status)
	assert.Contains(t, res.Message, "Failed to list services")

	// Test 3: No services
	mockReg.err = nil
	mockReg.services = []*configv1.UpstreamServiceConfig{}
	res = app.servicesHealthCheck(context.Background())
	assert.Equal(t, "ok", res.Status)
	assert.Contains(t, res.Message, "No services registered")

	// Test 4: Service skipped (Empty config)
	s1 := &configv1.UpstreamServiceConfig{
		Name: proto.String("s1"),
	}
	mockReg.services = []*configv1.UpstreamServiceConfig{s1}
	res = app.servicesHealthCheck(context.Background())
	assert.Equal(t, "ok", res.Status)
	assert.Contains(t, res.Message, "1 services healthy")

	// Test 5: Service fails (Invalid HTTP)
	// We expect doctor.CheckService to return Error for invalid URL/host
	s2 := &configv1.UpstreamServiceConfig{
		Name: proto.String("s2"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://invalid.local.domain"),
			},
		},
	}
	mockReg.services = []*configv1.UpstreamServiceConfig{s2}
	res = app.servicesHealthCheck(context.Background())
	assert.Equal(t, "degraded", res.Status)
	assert.Contains(t, res.Message, "s2")
	// Message might contain "dial tcp" or "lookup"
}
