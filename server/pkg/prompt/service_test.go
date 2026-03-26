// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPrompt ...
// Summary: MockPrompt
	mock.Mock
}

// Prompt ...
// Summary: Prompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	return args.Get(0).(*mcp.Prompt)
}

// Service ...
// Summary: Service
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	return args.String(0)
}

// Definition ...
// Summary: Definition
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*configv1.PromptDefinition)
}

// Get ...
// Summary: Get
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	ctx context.Context,
	args json.RawMessage,
) (*mcp.GetPromptResult, error) {
	calledArgs := m.Called(ctx, args)
	return calledArgs.Get(0).(*mcp.GetPromptResult), calledArgs.Error(1)
}

// MockErrorPrompt ...
// Summary: MockErrorPrompt
	MockPrompt
}

// Get ...
// Summary: Get
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	_ context.Context,
	_ json.RawMessage,
) (*mcp.GetPromptResult, error) {
	return nil, fmt.Errorf("error from Get")
}

// MockPromptManager ...
// Summary: MockPromptManager
	mock.Mock
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
	args := m.Called(name)
	return args.Get(0).(prompt.Prompt), args.Bool(1)
}

// AddPrompt ...
// Summary: AddPrompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(p)
}

// UpdatePrompt ...
// Summary: UpdatePrompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(p)
}

// RemovePrompt ...
// Summary: RemovePrompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(name)
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
	args := m.Called()
	return args.Get(0).([]prompt.Prompt)
}

// SetMCPServer ...
// Summary: SetMCPServer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(mcpServer)
}

// ClearPromptsForService ...
// Summary: ClearPromptsForService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(serviceID)
}

// GetServiceInfo ...
// Summary: GetServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(serviceID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}

// TestService_ListPrompts ...
// Summary: TestService_ListPrompts
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockPromptManager := new(MockPromptManager)
	service := prompt.NewService(mockPromptManager)

	mockPrompts := []prompt.Prompt{
		&MockPrompt{},
	}
	mockMCPPrompt := &mcp.Prompt{Name: "test_prompt"}

	// Configure the mock prompt to return the mock MCP prompt
	mockPrompts[0].(*MockPrompt).On("Prompt").Return(mockMCPPrompt)
	mockPromptManager.On("ListPrompts").Return(mockPrompts)

	result, err := service.ListPrompts(context.Background(), &mcp.ListPromptsRequest{})

	assert.NoError(t, err)
	assert.Len(t, result.Prompts, 1)
	assert.Equal(t, "test_prompt", result.Prompts[0].Name)
	mockPromptManager.AssertExpectations(t)
	mockPrompts[0].(*MockPrompt).AssertExpectations(t)
}

// TestMessage ...
// Summary: TestMessage
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// TestTextContent ...
// Summary: TestTextContent
	Type string `json:"type"`
	Text string `json:"text"`
}

// TestService_GetPrompt ...
// Summary: TestService_GetPrompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockPromptManager := new(MockPromptManager)
	service := prompt.NewService(mockPromptManager)

	mockPrompt := new(MockPrompt)
	mockPromptResult := &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: "Hello, world!",
				},
			},
		},
	}

	rawArgs := json.RawMessage(`{"key":"value"}`)
	mockPromptManager.On("GetPrompt", "test_prompt").Return(mockPrompt, true)
	mockPrompt.On("Get", context.Background(), rawArgs).Return(mockPromptResult, nil)

	result, err := service.GetPrompt(context.Background(), &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name: "test_prompt",
			Arguments: map[string]string{
				"key": "value",
			},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, mockPromptResult, result)
	mockPromptManager.AssertExpectations(t)
	mockPrompt.AssertExpectations(t)
}

// TestService_GetPrompt_NotFound ...
// Summary: TestService_GetPrompt_NotFound
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockPromptManager := new(MockPromptManager)
	service := prompt.NewService(mockPromptManager)

	mockPromptManager.On("GetPrompt", "not_found_prompt").Return((*MockPrompt)(nil), false)

	_, err := service.GetPrompt(context.Background(), &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name: "not_found_prompt",
		},
	})

	assert.ErrorIs(t, err, prompt.ErrPromptNotFound)
	mockPromptManager.AssertExpectations(t)
}

// TestService_GetPrompt_GetError ...
// Summary: TestService_GetPrompt_GetError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockPromptManager := new(MockPromptManager)
	service := prompt.NewService(mockPromptManager)

	mockPrompt := new(MockErrorPrompt)
	rawArgs := json.RawMessage(`{"key":"value"}`)
	mockPromptManager.On("GetPrompt", "test_prompt").Return(mockPrompt, true)
	mockPrompt.On("Get", context.Background(), rawArgs).Return(nil, fmt.Errorf("error from Get"))

	_, err := service.GetPrompt(context.Background(), &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name: "test_prompt",
			Arguments: map[string]string{
				"key": "value",
			},
		},
	})

	assert.Error(t, err)
}

// TestService_SetMCPServer ...
// Summary: TestService_SetMCPServer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockPromptManager := new(MockPromptManager)
	service := prompt.NewService(mockPromptManager)
	server := &mcp.Server{}

	mockPromptManager.On("SetMCPServer", mock.Anything).Return()

	service.SetMCPServer(server)

	mockPromptManager.AssertExpectations(t)
}
