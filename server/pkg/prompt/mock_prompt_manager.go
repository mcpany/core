// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"github.com/stretchr/testify/mock"
)

// MockPromptManager is a mock of ManagerInterface.
type MockPromptManager struct {
	mock.Mock
}

// AddPrompt is a mock method.
func (m *MockPromptManager) AddPrompt(p Prompt) {
	m.Called(p)
}

// UpdatePrompt is a mock method.
func (m *MockPromptManager) UpdatePrompt(p Prompt) {
	m.Called(p)
}

// GetPrompt is a mock method.
func (m *MockPromptManager) GetPrompt(name string) (Prompt, bool) {
	args := m.Called(name)
	return args.Get(0).(Prompt), args.Bool(1)
}

// ListPrompts is a mock method.
func (m *MockPromptManager) ListPrompts() []Prompt {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]Prompt)
}

// ClearPromptsForService is a mock method.
func (m *MockPromptManager) ClearPromptsForService(serviceID string) {
	m.Called(serviceID)
}

// SetMCPServer is a mock method.
func (m *MockPromptManager) SetMCPServer(mcpServer MCPServerProvider) {
	m.Called(mcpServer)
}
