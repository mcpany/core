package mcp

import (
	"context"
	"errors"
	"net/http"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClientSession is a mock implementation of the ClientSession interface
type MockClientSession struct {
	mock.Mock
}

func (m *MockClientSession) ListTools(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*mcp.ListToolsResult), args.Error(1)
}

func (m *MockClientSession) ListPrompts(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*mcp.ListPromptsResult), args.Error(1)
}

func (m *MockClientSession) ListResources(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*mcp.ListResourcesResult), args.Error(1)
}

func (m *MockClientSession) GetPrompt(ctx context.Context, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*mcp.GetPromptResult), args.Error(1)
}

func (m *MockClientSession) ReadResource(ctx context.Context, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*mcp.ReadResourceResult), args.Error(1)
}

func (m *MockClientSession) CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mcp.CallToolResult), args.Error(1)
}

func (m *MockClientSession) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestMcpConnection_CallTool(t *testing.T) {
	stdioConfig := configv1.McpStdioConnection_builder{}.Build()
	stdioConfig.SetCommand("echo")
	conn := &mcpConnection{
		stdioConfig: stdioConfig,
	}
	ctx := context.Background()
	params := &mcp.CallToolParams{Name: "test-tool"}

	t.Run("successful call", func(t *testing.T) {
		expectedResult := &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "success"}}}
		originalConnect := connectForTesting
		SetConnectForTesting(func(_ *mcp.Client, ctx context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
			mockSession := new(MockClientSession)
			mockSession.On("CallTool", ctx, params).Return(expectedResult, nil)
			mockSession.On("Close").Return(nil)
			return mockSession, nil
		})
		defer func() { connectForTesting = originalConnect }()

		result, err := conn.CallTool(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("connection error", func(t *testing.T) {
		connectErr := errors.New("connection failed")
		originalConnect := connectForTesting
		SetConnectForTesting(func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
			return nil, connectErr
		})
		defer func() { connectForTesting = originalConnect }()

		_, err := conn.CallTool(ctx, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), connectErr.Error())
	})

	t.Run("tool call error", func(t *testing.T) {
		toolErr := errors.New("tool call failed")
		originalConnect := connectForTesting
		SetConnectForTesting(func(_ *mcp.Client, ctx context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
			mockSession := new(MockClientSession)
			mockSession.On("CallTool", ctx, params).Return(nil, toolErr)
			mockSession.On("Close").Return(nil)
			return mockSession, nil
		})
		defer func() { connectForTesting = originalConnect }()

		_, err := conn.CallTool(ctx, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), toolErr.Error())
	})
}

func TestSetTestingHooks(t *testing.T) {
	t.Run("SetNewClientImplForTesting", func(t *testing.T) {
		var called bool
		SetNewClientImplForTesting(func(_ *mcp.Client, _ *configv1.McpStdioConnection, _ string, _ *http.Client) client.MCPClient {
			called = true
			return nil
		})
		assert.NotNil(t, newClientImplForTesting)
		newClientImplForTesting(nil, nil, "", nil)
		assert.True(t, called)
		newClientImplForTesting = nil // Reset for other tests
	})

	t.Run("SetNewClientForTesting", func(t *testing.T) {
		var called bool
		SetNewClientForTesting(func(_ *mcp.Implementation) *mcp.Client {
			called = true
			return nil
		})
		assert.NotNil(t, newClientForTesting)
		newClientForTesting(nil)
		assert.True(t, called)
		newClientForTesting = nil // Reset for other tests
	})

	t.Run("SetConnectForTesting", func(t *testing.T) {
		var called bool
		SetConnectForTesting(func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
			called = true
			return nil, nil
		})
		assert.NotNil(t, connectForTesting)
		_, _ = connectForTesting(nil, context.Background(), nil, nil)
		assert.True(t, called)
		connectForTesting = nil // Reset for other tests
	})
}

func TestMcpPrompt_Service(t *testing.T) {
	prompt := &mcpPrompt{service: "test-service"}
	assert.Equal(t, "test-service", prompt.Service())
}

func TestMcpResource_Service(t *testing.T) {
	resource := &mcpResource{service: "test-service"}
	assert.Equal(t, "test-service", resource.Service())
}

func TestMcpResource_Subscribe(t *testing.T) {
	resource := &mcpResource{}
	err := resource.Subscribe(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}
