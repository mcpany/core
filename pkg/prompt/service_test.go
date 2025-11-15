/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package prompt

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mcpany/core/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockErrorPrompt struct {
	MockPrompt
}

func (m *MockErrorPrompt) Get(
	ctx context.Context,
	args json.RawMessage,
) (*mcp.GetPromptResult, error) {
	return nil, fmt.Errorf("error from Get")
}

type MockPromptManager struct {
	mock.Mock
}

func (m *MockPromptManager) GetPrompt(name string) (Prompt, bool) {
	args := m.Called(name)
	return args.Get(0).(Prompt), args.Bool(1)
}

func (m *MockPromptManager) AddPrompt(p Prompt) {
	m.Called(p)
}

func (m *MockPromptManager) RemovePrompt(name string) {
	m.Called(name)
}

func (m *MockPromptManager) ListPrompts() []Prompt {
	args := m.Called()
	return args.Get(0).([]Prompt)
}

func (m *MockPromptManager) SetMCPServer(mcpServer MCPServerProvider) {
	m.Called(mcpServer)
}

func (m *MockPromptManager) ClearPromptsForService(serviceID string) {
	m.Called(serviceID)
}

func (m *MockPromptManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}

func TestService_ListPrompts(t *testing.T) {
	mockPromptManager := new(MockPromptManager)
	service := NewService(mockPromptManager)

	mockPrompts := []Prompt{
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

type TestMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type TestTextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func TestService_GetPrompt(t *testing.T) {
	mockPromptManager := new(MockPromptManager)
	service := NewService(mockPromptManager)

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

func TestService_GetPrompt_NotFound(t *testing.T) {
	mockPromptManager := new(MockPromptManager)
	service := NewService(mockPromptManager)

	mockPromptManager.On("GetPrompt", "not_found_prompt").Return((*MockPrompt)(nil), false)

	_, err := service.GetPrompt(context.Background(), &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name: "not_found_prompt",
		},
	})

	assert.ErrorIs(t, err, ErrPromptNotFound)
	mockPromptManager.AssertExpectations(t)
}

func TestService_GetPrompt_GetError(t *testing.T) {
	mockPromptManager := new(MockPromptManager)
	service := NewService(mockPromptManager)

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

func TestService_SetMCPServer(t *testing.T) {
	mockPromptManager := new(MockPromptManager)
	service := NewService(mockPromptManager)
	mockMCPServer := &mcp.Server{}
	provider := NewMCPServerProvider(mockMCPServer)

	mockPromptManager.On("SetMCPServer", provider).Return()
	service.SetMCPServer(mockMCPServer)

	mockPromptManager.AssertExpectations(t)
}
