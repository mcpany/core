package serviceregistry

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_routerv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type mockTool struct {
	tool *mcp_routerv1.Tool
}

func (m *mockTool) Tool() *mcp_routerv1.Tool {
	return m.tool
}

func (m *mockTool) Execute(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

func (m *mockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func (m *mockTool) MCPTool() *mcp.Tool {
	t, _ := tool.ConvertProtoToMCPTool(m.tool)
	return t
}

type threadSafeToolManager struct {
	tool.ManagerInterface
	mu    sync.RWMutex
	tools map[string]tool.Tool
}

func newThreadSafeToolManager() *threadSafeToolManager {
	return &threadSafeToolManager{
		tools: make(map[string]tool.Tool),
	}
}

func (m *threadSafeToolManager) AddTool(t tool.Tool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tools[t.Tool().GetName()] = t
	return nil
}

func (m *threadSafeToolManager) ClearToolsForService(serviceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, t := range m.tools {
		if t.Tool().GetServiceId() == serviceID {
			delete(m.tools, name)
		}
	}
}

func (m *threadSafeToolManager) GetTool(name string) (tool.Tool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tools[name]
	return t, ok
}

func (m *threadSafeToolManager) ListTools() []tool.Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	tools := make([]tool.Tool, 0, len(m.tools))
	for _, t := range m.tools {
		tools = append(tools, t)
	}
	return tools
}

func TestServiceRegistry_RegisterService_DuplicateNameDoesNotClearExisting(t *testing.T) {
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					serviceID, err := util.SanitizeServiceName(serviceName)
					require.NoError(t, err)
					return serviceID, nil, nil, nil
				},
			}, nil
		},
	}
	tm := newThreadSafeToolManager()
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	// Register the first service with a tool
	serviceConfig1 := &configv1.UpstreamServiceConfig{}
	serviceConfig1.SetName("test-service")
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig1)
	require.NoError(t, err, "First registration should succeed")

	// Add a tool to the service
	tool1 := &mockTool{tool: &mcp_routerv1.Tool{Name: proto.String("tool1"), ServiceId: proto.String(serviceID)}}
	err = tm.AddTool(tool1)
	require.NoError(t, err)

	// Verify the tool is there
	_, ok := tm.GetTool("tool1")
	assert.True(t, ok, "Tool should be present after first registration")

	// Attempt to register another service with the same name
	serviceConfig2 := &configv1.UpstreamServiceConfig{}
	serviceConfig2.SetName("test-service")
	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig2)
	require.Error(t, err, "Second registration with the same name should fail")

	// Verify that the tool from the first service is still there
	_, ok = tm.GetTool("tool1")
	assert.True(t, ok, "Tool should still be present after failed duplicate registration")
}

// mockPrompt is a mock implementation of prompt.Prompt for testing.
type mockPrompt struct {
	p         *mcp.Prompt
	serviceID string
}

func (m *mockPrompt) Prompt() *mcp.Prompt {
	return m.p
}

func (m *mockPrompt) Service() string {
	return m.serviceID
}

func (m *mockPrompt) Get(_ context.Context, _ json.RawMessage) (*mcp.GetPromptResult, error) {
	return nil, nil
}

// mockResource is a mock implementation of resource.Resource for testing.
type mockResource struct {
	r         *mcp.Resource
	serviceID string
}

func (m *mockResource) Resource() *mcp.Resource {
	return m.r
}

func (m *mockResource) Service() string {
	return m.serviceID
}

func (m *mockResource) Read(_ context.Context) (*mcp.ReadResourceResult, error) {
	return nil, nil
}

func (m *mockResource) Subscribe(_ context.Context) error {
	return nil
}

func TestServiceRegistry_UnregisterService_ClearsAllData(t *testing.T) {
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					serviceID, err := util.SanitizeServiceName(serviceName)
					require.NoError(t, err)
					return serviceID, nil, nil, nil
				},
			}, nil
		},
	}
	tm := newThreadSafeToolManager()
	pm := prompt.NewManager()
	rm := resource.NewManager()
	registry := New(f, tm, pm, rm, auth.NewManager())

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err, "Registration should succeed")

	// Manually add items to the managers
	err = tm.AddTool(&mockTool{tool: &mcp_routerv1.Tool{Name: proto.String("test-tool"), ServiceId: proto.String(serviceID)}})
	require.NoError(t, err)
	pm.AddPrompt(&mockPrompt{serviceID: serviceID, p: &mcp.Prompt{Name: "test-prompt"}})
	rm.AddResource(&mockResource{serviceID: serviceID, r: &mcp.Resource{URI: "test-resource"}})

	// Verify that the data is there
	assert.NotEmpty(t, tm.ListTools())
	assert.NotEmpty(t, pm.ListPrompts())
	assert.NotEmpty(t, rm.ListResources())

	// Unregister the service
	err = registry.UnregisterService(context.Background(), serviceID)
	require.NoError(t, err, "Unregistration should succeed")

	// Verify that all data has been cleared
	assert.Empty(t, tm.ListTools(), "Tools should be cleared after unregistration")
	assert.Empty(t, pm.ListPrompts(), "Prompts should be cleared after unregistration")
	assert.Empty(t, rm.ListResources(), "Resources should be cleared after unregistration")
}

func TestServiceRegistry_UnregisterService_CallsShutdown(t *testing.T) {
	shutdownCalled := false
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					serviceID, err := util.SanitizeServiceName(serviceName)
					require.NoError(t, err)
					return serviceID, nil, nil, nil
				},
				shutdownFunc: func() error {
					shutdownCalled = true
					return nil
				},
			}, nil
		},
	}
	tm := newThreadSafeToolManager()
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err, "Registration should succeed")

	// Unregister the service
	err = registry.UnregisterService(context.Background(), serviceID)
	require.NoError(t, err, "Unregistration should succeed")

	// Verify that the shutdown method was called
	assert.True(t, shutdownCalled, "Shutdown method should be called on unregister")
}
