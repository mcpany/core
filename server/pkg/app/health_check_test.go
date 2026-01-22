// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

type mockServiceRegistry struct {
	services []*configv1.UpstreamServiceConfig
	errors   map[string]string
}

func (m *mockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	return "", nil, nil, nil
}
func (m *mockServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	return nil
}
func (m *mockServiceRegistry) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	return m.services, nil
}
func (m *mockServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return nil, false
}
func (m *mockServiceRegistry) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	return nil, false
}
func (m *mockServiceRegistry) GetServiceError(serviceID string) (string, bool) {
	err, ok := m.errors[serviceID]
	return err, ok
}
func (m *mockServiceRegistry) GetServiceHealth(serviceID string) (string, time.Duration, time.Time, bool) {
	// Simple mock implementation: if error exists, return it.
	if err, ok := m.errors[serviceID]; ok {
		return err, 0, time.Time{}, true
	}
	// Assume healthy if in services list
	for _, s := range m.services {
		if s.GetId() == serviceID {
			return "", 0, time.Time{}, true
		}
	}
	return "", 0, time.Time{}, false
}
func (m *mockServiceRegistry) Close(ctx context.Context) error {
	return nil
}

func TestServicesHealthCheck(t *testing.T) {
	// Setup
	registry := &mockServiceRegistry{
		services: []*configv1.UpstreamServiceConfig{
			{Name: proto.String("healthy_service"), Id: proto.String("healthy_service")},
			{Name: proto.String("failing_service"), Id: proto.String("failing_service")},
		},
		errors: map[string]string{
			"failing_service": "connection refused",
		},
	}

	app := &Application{
		ServiceRegistry: registry,
	}

	// Test
	result := app.servicesHealthCheck(context.Background())

	// Verify
	assert.Equal(t, "degraded", result.Status)
	assert.Contains(t, result.Message, "failing_service")
	assert.Contains(t, result.Message, "connection refused")
	assert.NotContains(t, result.Message, "healthy_service")
}

func TestServicesHealthCheck_AllHealthy(t *testing.T) {
	// Setup
	registry := &mockServiceRegistry{
		services: []*configv1.UpstreamServiceConfig{
			{Name: proto.String("healthy_service"), Id: proto.String("healthy_service")},
		},
		errors: map[string]string{},
	}

	app := &Application{
		ServiceRegistry: registry,
	}

	// Test
	result := app.servicesHealthCheck(context.Background())

	// Verify
	assert.Equal(t, "ok", result.Status)
	assert.Empty(t, result.Message)
}
