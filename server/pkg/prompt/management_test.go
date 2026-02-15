package prompt

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockServiceRegistry is a mock implementation of the ServiceRegistryInterface.
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
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}

// MockPrompt is a mock implementation of the Prompt interface.
type MockPrompt struct {
	mock.Mock
}

func (m *MockPrompt) Prompt() *mcp.Prompt {
	args := m.Called()
	return args.Get(0).(*mcp.Prompt)
}

func (m *MockPrompt) Service() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	calledArgs := m.Called(ctx, args)
	return calledArgs.Get(0).(*mcp.GetPromptResult), calledArgs.Error(1)
}

func TestPromptManager(t *testing.T) {
	promptManager := NewManager()

	t.Run("add and get prompt", func(t *testing.T) {
		mockPrompt := new(MockPrompt)
		mcpPrompt := &mcp.Prompt{Name: "test-prompt"}
		mockPrompt.On("Prompt").Return(mcpPrompt)
		promptManager.AddPrompt(mockPrompt)

		p, ok := promptManager.GetPrompt("test-prompt")
		assert.True(t, ok)
		assert.Equal(t, mockPrompt, p)

		_, ok = promptManager.GetPrompt("non-existent")
		assert.False(t, ok)
	})

	t.Run("list prompts", func(t *testing.T) {
		// Clear existing prompts
		promptManager.prompts.Clear()

		mockPrompt1 := new(MockPrompt)
		mockPrompt1.On("Prompt").Return(&mcp.Prompt{Name: "prompt1"})
		promptManager.AddPrompt(mockPrompt1)

		mockPrompt2 := new(MockPrompt)
		mockPrompt2.On("Prompt").Return(&mcp.Prompt{Name: "prompt2"})
		promptManager.AddPrompt(mockPrompt2)

		prompts := promptManager.ListPrompts()
		assert.Len(t, prompts, 2)
	})

	t.Run("clear prompts for service", func(t *testing.T) {
		// Clear existing prompts
		promptManager.prompts.Clear()

		mockPrompt1 := new(MockPrompt)
		mockPrompt1.On("Prompt").Return(&mcp.Prompt{Name: "service1.prompt1"})
		mockPrompt1.On("Service").Return("service1")
		promptManager.AddPrompt(mockPrompt1)

		mockPrompt2 := new(MockPrompt)
		mockPrompt2.On("Prompt").Return(&mcp.Prompt{Name: "service2.prompt2"})
		mockPrompt2.On("Service").Return("service2")
		promptManager.AddPrompt(mockPrompt2)

		promptManager.ClearPromptsForService("service1")
		prompts := promptManager.ListPrompts()
		assert.Len(t, prompts, 1)
		assert.Equal(t, "service2.prompt2", prompts[0].Prompt().Name)
	})

	t.Run("set mcp server", func(t *testing.T) {
		server := &mcp.Server{}
		provider := NewMCPServerProvider(server)

		promptManager.SetMCPServer(provider)

		assert.Equal(t, provider, promptManager.mcpServer)
	})

	t.Run("add duplicate prompt", func(t *testing.T) {
		// Clear existing prompts
		promptManager.prompts.Clear()

		mockPrompt1 := new(MockPrompt)
		mockPrompt1.On("Prompt").Return(&mcp.Prompt{Name: "duplicate-prompt"})
		mockPrompt1.On("Service").Return("service1")
		promptManager.AddPrompt(mockPrompt1)

		mockPrompt2 := new(MockPrompt)
		mockPrompt2.On("Prompt").Return(&mcp.Prompt{Name: "duplicate-prompt"})
		mockPrompt2.On("Service").Return("service2")
		promptManager.AddPrompt(mockPrompt2)

		p, ok := promptManager.GetPrompt("duplicate-prompt")
		assert.True(t, ok)
		assert.Equal(t, "service2", p.Service())
	})

	t.Run("update prompt", func(t *testing.T) {
		promptManager.prompts.Clear()

		mockPrompt1 := new(MockPrompt)
		mockPrompt1.On("Prompt").Return(&mcp.Prompt{Name: "prompt1"})
		promptManager.AddPrompt(mockPrompt1)

		mockPrompt2 := new(MockPrompt)
		mockPrompt2.On("Prompt").Return(&mcp.Prompt{Name: "prompt1"})
		promptManager.UpdatePrompt(mockPrompt2)

		p, ok := promptManager.GetPrompt("prompt1")
		assert.True(t, ok)
		assert.Equal(t, mockPrompt2, p)
	})
}
