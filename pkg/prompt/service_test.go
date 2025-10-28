/*
 * Copyright 2025 Author(s) of MCP-XY
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

package prompt_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpxy/core/pkg/prompt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func (m *MockPrompt) Get(
	ctx context.Context,
	args json.RawMessage,
) (*mcp.GetPromptResult, error) {
	calledArgs := m.Called(ctx, args)
	return calledArgs.Get(0).(*mcp.GetPromptResult), calledArgs.Error(1)
}

type MockPromptManager struct {
	mock.Mock
}

func (m *MockPromptManager) GetPrompt(name string) (prompt.Prompt, bool) {
	args := m.Called(name)
	return args.Get(0).(prompt.Prompt), args.Bool(1)
}

func (m *MockPromptManager) AddPrompt(p prompt.Prompt) {
	m.Called(p)
}

func (m *MockPromptManager) RemovePrompt(name string) {
	m.Called(name)
}

func (m *MockPromptManager) ListPrompts() []prompt.Prompt {
	args := m.Called()
	return args.Get(0).([]prompt.Prompt)
}

func (m *MockPromptManager) OnListChanged(f func()) {
	m.Called(f)
}

func TestService_ListPrompts(t *testing.T) {
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

func TestService_GetPrompt_NotFound(t *testing.T) {
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
