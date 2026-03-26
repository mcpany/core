// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
// MockClientSession is a mock implementation of the ClientSession interface
// Summary: MockClientSession
	mock.Mock
}

// ListTools ...
// Summary: ListTools
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, params)
	return args.Get(0).(*mcp.ListToolsResult), args.Error(1)
}

// ListPrompts ...
// Summary: ListPrompts
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, params)
	return args.Get(0).(*mcp.ListPromptsResult), args.Error(1)
}

// ListResources ...
// Summary: ListResources
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, params)
	return args.Get(0).(*mcp.ListResourcesResult), args.Error(1)
}

// GetPrompt ...
// Summary: GetPrompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, params)
	return args.Get(0).(*mcp.GetPromptResult), args.Error(1)
}

// ReadResource ...
// Summary: ReadResource
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, params)
	return args.Get(0).(*mcp.ReadResourceResult), args.Error(1)
}

// CallTool ...
// Summary: CallTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mcp.CallToolResult), args.Error(1)
}

// Close ...
// Summary: Close
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	return args.Error(0)
}

// TestMcpConnection_CallTool ...
// Summary: TestMcpConnection_CallTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestSetTestingHooks ...
// Summary: TestSetTestingHooks
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestMcpPrompt_Service ...
// Summary: TestMcpPrompt_Service
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	prompt := &mcpPrompt{service: "test-service"}
	assert.Equal(t, "test-service", prompt.Service())
}

// TestMcpResource_Service ...
// Summary: TestMcpResource_Service
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	resource := &mcpResource{service: "test-service"}
	assert.Equal(t, "test-service", resource.Service())
}

// TestMcpResource_Subscribe ...
// Summary: TestMcpResource_Subscribe
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	resource := &mcpResource{}
	err := resource.Subscribe(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}
