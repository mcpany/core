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

package prompt_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/prompt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPromptManager struct {
	mock.Mock
}

func (m *MockPromptManager) GetPrompt(name string) (*prompt.Prompt, bool) {
	args := m.Called(name)
	return args.Get(0).(*prompt.Prompt), args.Bool(1)
}

func (m *MockPromptManager) ListPrompts() []*prompt.Prompt {
	args := m.Called()
	return args.Get(0).([]*prompt.Prompt)
}

func (m *MockPromptManager) ClearPromptsForService(serviceID string) {
	m.Called(serviceID)
}

func (m *MockPromptManager) AddPrompt(p *prompt.Prompt) {
	m.Called(p)
}

func TestService_ListPrompts(t *testing.T) {
	mockPromptManager := new(MockPromptManager)
	service := prompt.NewService(mockPromptManager)

	mockPrompts := []*prompt.Prompt{
		{
			Name:        "test_prompt",
			Description: "A test prompt",
		},
	}
	mockPromptManager.On("ListPrompts").Return(mockPrompts)

	result, err := service.ListPrompts(context.Background(), &mcp.ListPromptsRequest{})

	assert.NoError(t, err)
	assert.Len(t, result.Prompts, 1)
	assert.Equal(t, "test_prompt", result.Prompts[0].Name)
	mockPromptManager.AssertExpectations(t)
}
