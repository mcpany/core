// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestMockTool implements tool.Tool interface
type TestMockTool struct {
	toolDef *v1.Tool
}

func (m *TestMockTool) Tool() *v1.Tool { return m.toolDef }
func (m *TestMockTool) MCPTool() *mcp.Tool { return nil }
func (m *TestMockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) { return nil, nil }
func (m *TestMockTool) GetCacheConfig() *configv1.CacheConfig { return nil }

// MockServiceStore implements config.ServiceStore
type MockServiceStore struct {
	services []*configv1.UpstreamServiceConfig
}

func (s *MockServiceStore) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) { return nil, nil }
func (s *MockServiceStore) HasConfigSources() bool { return false }
func (s *MockServiceStore) SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error { return nil }
func (s *MockServiceStore) GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) { return nil, nil }
func (s *MockServiceStore) ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) { return s.services, nil }
func (s *MockServiceStore) DeleteService(ctx context.Context, name string) error { return nil }
func (s *MockServiceStore) ListSecrets(ctx context.Context) ([]*configv1.Secret, error) { return nil, nil }
func (s *MockServiceStore) SaveSecret(ctx context.Context, secret *configv1.Secret) error { return nil }
func (s *MockServiceStore) GetSecret(ctx context.Context, id string) (*configv1.Secret, error) { return nil, nil }
func (s *MockServiceStore) DeleteSecret(ctx context.Context, id string) error { return nil }
// Add other required methods stubbed
func (s *MockServiceStore) ListProfiles(ctx context.Context) ([]*configv1.ProfileDefinition, error) { return nil, nil }
func (s *MockServiceStore) SaveProfile(ctx context.Context, p *configv1.ProfileDefinition) error { return nil }
func (s *MockServiceStore) GetProfile(ctx context.Context, name string) (*configv1.ProfileDefinition, error) { return nil, nil }
func (s *MockServiceStore) DeleteProfile(ctx context.Context, name string) error { return nil }
func (s *MockServiceStore) ListServiceCollections(ctx context.Context) ([]*configv1.Collection, error) { return nil, nil }
func (s *MockServiceStore) SaveServiceCollection(ctx context.Context, c *configv1.Collection) error { return nil }
func (s *MockServiceStore) GetServiceCollection(ctx context.Context, name string) (*configv1.Collection, error) { return nil, nil }
func (s *MockServiceStore) DeleteServiceCollection(ctx context.Context, name string) error { return nil }
func (s *MockServiceStore) GetGlobalSettings(ctx context.Context) (*configv1.GlobalSettings, error) { return nil, nil }
func (s *MockServiceStore) SaveGlobalSettings(ctx context.Context, gs *configv1.GlobalSettings) error { return nil }
func (s *MockServiceStore) Close() error { return nil }
func (s *MockServiceStore) CreateUser(ctx context.Context, user *configv1.User) error { return nil }
func (s *MockServiceStore) GetUser(ctx context.Context, id string) (*configv1.User, error) { return nil, nil }
func (s *MockServiceStore) ListUsers(ctx context.Context) ([]*configv1.User, error) { return nil, nil }
func (s *MockServiceStore) UpdateUser(ctx context.Context, user *configv1.User) error { return nil }
func (s *MockServiceStore) DeleteUser(ctx context.Context, id string) error { return nil }
func (s *MockServiceStore) SaveToken(ctx context.Context, token *configv1.UserToken) error { return nil }
func (s *MockServiceStore) GetToken(ctx context.Context, userID, serviceID string) (*configv1.UserToken, error) { return nil, nil }
func (s *MockServiceStore) DeleteToken(ctx context.Context, userID, serviceID string) error { return nil }
func (s *MockServiceStore) ListCredentials(ctx context.Context) ([]*configv1.Credential, error) { return nil, nil }
func (s *MockServiceStore) GetCredential(ctx context.Context, id string) (*configv1.Credential, error) { return nil, nil }
func (s *MockServiceStore) SaveCredential(ctx context.Context, cred *configv1.Credential) error { return nil }
func (s *MockServiceStore) DeleteCredential(ctx context.Context, id string) error { return nil }

func TestHandleServices_ToolCount(t *testing.T) {
	// Setup
	busProvider, _ := bus.NewProvider(nil)
	tm := tool.NewManager(busProvider)

	// Add some tools
	// service-1 has 2 tools
	tm.AddTool(&TestMockTool{toolDef: &v1.Tool{Name: proto.String("tool1"), ServiceId: proto.String("service-1")}})
	tm.AddTool(&TestMockTool{toolDef: &v1.Tool{Name: proto.String("tool2"), ServiceId: proto.String("service-1")}})
	// service-2 has 1 tool
	tm.AddTool(&TestMockTool{toolDef: &v1.Tool{Name: proto.String("tool3"), ServiceId: proto.String("service-2")}})
	// service-3 has 0 tools

	app := NewApplication()
	app.ToolManager = tm

	// I'll create a dummy implementation of ServiceRegistryInterface
	app.ServiceRegistry = &TestMockServiceRegistry{
		services: []*configv1.UpstreamServiceConfig{
			{Id: proto.String("service-1"), Name: proto.String("service-1")},
			{Id: proto.String("service-2"), Name: proto.String("service-2")},
			{Id: proto.String("service-3"), Name: proto.String("service-3")},
		},
	}

	// Prepare request
	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	rr := httptest.NewRecorder()

	// The store argument is used as fallback if ServiceRegistry is nil, OR if GetAllServices returns error?
	// The code says:
	// if a.ServiceRegistry != nil { services = a.ServiceRegistry.GetAllServices() } else { services = store.ListServices() }
	// So passing nil store is risky if I rely on fallback, but here I rely on ServiceRegistry.
	// But `handleServices` signature requires `storage.Storage` (which includes ServiceStore).
	// I can pass nil and hope it doesn't crash if ServiceRegistry is present?
	// No, the function signature `handleServices(store storage.Storage)` is defined.
	// But `store` is only used in fallback.
	// Let's pass a dummy store just in case.

	handler := app.handleServices(&MockServiceStore{})
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var response []map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Len(t, response, 3)

	// Sort or find by name to verify
	svc1 := findService(response, "service-1")
	require.NotNil(t, svc1)
	assert.Equal(t, float64(2), svc1["tool_count"], "service-1 should have 2 tools")

	svc2 := findService(response, "service-2")
	require.NotNil(t, svc2)
	assert.Equal(t, float64(1), svc2["tool_count"], "service-2 should have 1 tool")

	svc3 := findService(response, "service-3")
	require.NotNil(t, svc3)
	assert.Equal(t, float64(0), svc3["tool_count"], "service-3 should have 0 tools")
}

func findService(list []map[string]any, name string) map[string]any {
	for _, s := range list {
		if s["name"] == name {
			return s
		}
	}
	return nil
}

// TestMockServiceRegistry implements serviceregistry.ServiceRegistryInterface
type TestMockServiceRegistry struct {
	services []*configv1.UpstreamServiceConfig
}

func (m *TestMockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) { return "", nil, nil, nil }
func (m *TestMockServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error { return nil }
func (m *TestMockServiceRegistry) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) { return m.services, nil }
func (m *TestMockServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) { return nil, false }
func (m *TestMockServiceRegistry) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) { return nil, false }
func (m *TestMockServiceRegistry) GetServiceError(serviceID string) (string, bool) { return "", false }
