// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// MockServiceRegistry for testing reconcileServices
type mockServiceRegistry struct {
	services     map[string]*configv1.UpstreamServiceConfig
	registered   []*configv1.UpstreamServiceConfig
	unregistered []string
	errors       map[string]string
}

func newMockServiceRegistry() *mockServiceRegistry {
	return &mockServiceRegistry{
		services: make(map[string]*configv1.UpstreamServiceConfig),
		errors:   make(map[string]string),
	}
}

func (m *mockServiceRegistry) RegisterService(ctx context.Context, config *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	m.registered = append(m.registered, config)
	m.services[config.GetName()] = config
	return config.GetName(), nil, nil, nil
}

func (m *mockServiceRegistry) UnregisterService(ctx context.Context, name string) error {
	m.unregistered = append(m.unregistered, name)
	delete(m.services, name)
	return nil
}

func (m *mockServiceRegistry) GetService(name string) (*configv1.UpstreamServiceConfig, bool) {
	svc, ok := m.services[name]
	return svc, ok
}

func (m *mockServiceRegistry) GetServiceConfig(name string) (*configv1.UpstreamServiceConfig, bool) {
	svc, ok := m.services[name]
	return svc, ok
}

func (m *mockServiceRegistry) GetServiceInfo(name string) (*tool.ServiceInfo, bool) {
	return nil, false
}

func (m *mockServiceRegistry) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	var list []*configv1.UpstreamServiceConfig
	for _, svc := range m.services {
		list = append(list, svc)
	}
	return list, nil
}

func (m *mockServiceRegistry) UpdateServiceError(name string, errMsg string) {
	m.errors[name] = errMsg
}

func (m *mockServiceRegistry) GetServiceError(name string) (string, bool) {
	err, ok := m.errors[name]
	return err, ok
}

func (m *mockServiceRegistry) ClearServiceError(name string) {
	delete(m.errors, name)
}

func (m *mockServiceRegistry) GetAllServiceErrors() map[string]string {
	return m.errors
}

func TestReconcileServices(t *testing.T) {
	tests := []struct {
		name          string
		initial       []*configv1.UpstreamServiceConfig
		config        *configv1.McpAnyServerConfig
		expectReg     int
		expectUnreg   int
		expectNames   []string // Names of services expected to be registered
		expectRemoved []string // Names of services expected to be unregistered
	}{
		{
			name: "no changes",
			initial: []*configv1.UpstreamServiceConfig{
				configv1.UpstreamServiceConfig_builder{Name: proto.String("svc1"), Id: proto.String("id1"), SanitizedName: proto.String("svc1")}.Build(),
			},
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{Name: proto.String("svc1"), Id: proto.String("id1"), SanitizedName: proto.String("svc1")}.Build(),
				},
				GlobalSettings: configv1.GlobalSettings_builder{AutoDiscoverLocal: proto.Bool(false)}.Build(),
			}.Build(),
			expectReg:   0,
			expectUnreg: 0,
		},
		{
			name: "add new service",
			initial: []*configv1.UpstreamServiceConfig{
				configv1.UpstreamServiceConfig_builder{Name: proto.String("svc1"), Id: proto.String("id1"), SanitizedName: proto.String("svc1")}.Build(),
			},
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{Name: proto.String("svc1"), Id: proto.String("id1"), SanitizedName: proto.String("svc1")}.Build(),
					configv1.UpstreamServiceConfig_builder{Name: proto.String("svc2"), Id: proto.String("id2"), SanitizedName: proto.String("svc2")}.Build(),
				},
				GlobalSettings: configv1.GlobalSettings_builder{AutoDiscoverLocal: proto.Bool(false)}.Build(),
			}.Build(),
			expectReg:   1,
			expectUnreg: 0,
			expectNames: []string{"svc2"},
		},
		{
			name: "remove service",
			initial: []*configv1.UpstreamServiceConfig{
				configv1.UpstreamServiceConfig_builder{Name: proto.String("svc1"), Id: proto.String("id1"), SanitizedName: proto.String("svc1")}.Build(),
				configv1.UpstreamServiceConfig_builder{Name: proto.String("svc2"), Id: proto.String("id2"), SanitizedName: proto.String("svc2")}.Build(),
			},
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{Name: proto.String("svc1"), Id: proto.String("id1"), SanitizedName: proto.String("svc1")}.Build(),
				},
				GlobalSettings: configv1.GlobalSettings_builder{AutoDiscoverLocal: proto.Bool(false)}.Build(),
			}.Build(),
			expectReg:     0,
			expectUnreg:   1,
			expectRemoved: []string{"svc2"},
		},
		{
			name: "update service",
			initial: []*configv1.UpstreamServiceConfig{
				configv1.UpstreamServiceConfig_builder{Name: proto.String("svc1"), Id: proto.String("id1"), SanitizedName: proto.String("svc1"), Version: proto.String("old")}.Build(),
			},
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{Name: proto.String("svc1"), Id: proto.String("id1"), SanitizedName: proto.String("svc1"), Version: proto.String("new")}.Build(),
				},
				GlobalSettings: configv1.GlobalSettings_builder{AutoDiscoverLocal: proto.Bool(false)}.Build(),
			}.Build(),
			expectReg:     1,
			expectUnreg:   1,
			expectNames:   []string{"svc1"},
			expectRemoved: []string{"svc1"},
		},
		{
			name: "disable service acts as removal",
			initial: []*configv1.UpstreamServiceConfig{
				configv1.UpstreamServiceConfig_builder{Name: proto.String("svc1"), Id: proto.String("id1"), SanitizedName: proto.String("svc1")}.Build(),
			},
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{Name: proto.String("svc1"), Id: proto.String("id1"), SanitizedName: proto.String("svc1"), Disable: proto.Bool(true)}.Build(),
				},
				GlobalSettings: configv1.GlobalSettings_builder{AutoDiscoverLocal: proto.Bool(false)}.Build(),
			}.Build(),
			expectReg:     0,
			expectUnreg:   1,
			expectRemoved: []string{"svc1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := NewApplication()
			registry := newMockServiceRegistry()

			// Setup initial state
			for _, svc := range tc.initial {
				// Avoid mutating test cases directly by cloning
				clonedSvc := proto.Clone(svc).(*configv1.UpstreamServiceConfig)
				registry.services[clonedSvc.GetName()] = clonedSvc
			}

			app.ServiceRegistry = registry
			app.ToolManager = tool.NewManager(nil) // Needed for reload complete log

			app.reconcileServices(context.Background(), tc.config)

			var registeredNames []string
			for _, svc := range registry.registered {
				registeredNames = append(registeredNames, svc.GetName())
			}

			assert.Equal(t, tc.expectReg, len(registry.registered), "registered count mismatch, got: %v", registeredNames)
			assert.Equal(t, tc.expectUnreg, len(registry.unregistered), "unregistered count mismatch, got: %v", registry.unregistered)

			if len(tc.expectNames) > 0 {
				assert.ElementsMatch(t, tc.expectNames, registeredNames)
			} else {
				assert.Empty(t, registeredNames)
			}

			if len(tc.expectRemoved) > 0 {
				assert.ElementsMatch(t, tc.expectRemoved, registry.unregistered)
			} else {
				assert.Empty(t, registry.unregistered)
			}
		})
	}
}
