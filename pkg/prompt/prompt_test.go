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
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMCPServer is a mock implementation of the mcp.Server for testing.
type mockMCPServer struct {
	prompts               map[string]mcp.PromptHandler
	addPromptCalls        []*mcp.Prompt
	removePromptsCalls    []string
	removePromptsEmbedded []string
}

func newMockMCPServer() *mockMCPServer {
	return &mockMCPServer{
		prompts: make(map[string]mcp.PromptHandler),
	}
}

func (s *mockMCPServer) AddPrompt(prompt *mcp.Prompt, handler mcp.PromptHandler) {
	s.prompts[prompt.Name] = handler
	s.addPromptCalls = append(s.addPromptCalls, prompt)
}

func (s *mockMCPServer) RemovePrompts(names ...string) {
	for _, name := range names {
		delete(s.prompts, name)
		s.removePromptsCalls = append(s.removePromptsCalls, name)
	}
}

// mockMCPServerProvider is a mock implementation of the MCPServerProvider
// interface for testing.
type mockMCPServerProvider struct {
	server *mockMCPServer
}

func newMockMCPServerProvider() *mockMCPServerProvider {
	return &mockMCPServerProvider{
		server: newMockMCPServer(),
	}
}

func (p *mockMCPServerProvider) Server() mcpServer {
	return p.server
}

// mockPrompt is a mock implementation of the Prompt interface for testing.
type mockPrompt struct {
	name    string
	service string
}

func (p *mockPrompt) Prompt() *mcp.Prompt {
	return &mcp.Prompt{Name: p.name}
}

func (p *mockPrompt) Service() string {
	return p.service
}

func (p *mockPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{}, nil
}

func TestNewPromptManager(t *testing.T) {
	pm := NewPromptManager()
	assert.NotNil(t, pm)
	assert.NotNil(t, pm.prompts)
}

func TestPromptManager_AddGetListRemovePrompt(t *testing.T) {
	pm := NewPromptManager()
	prompt1 := &mockPrompt{name: "prompt1", service: "service1"}
	prompt2 := &mockPrompt{name: "prompt2", service: "service2"}

	// Add
	pm.AddPrompt(prompt1)
	pm.AddPrompt(prompt2)

	// Get
	p, ok := pm.GetPrompt("prompt1")
	require.True(t, ok)
	assert.Equal(t, prompt1, p)

	p, ok = pm.GetPrompt("prompt2")
	require.True(t, ok)
	assert.Equal(t, prompt2, p)

	_, ok = pm.GetPrompt("non-existent")
	assert.False(t, ok)

	// List
	prompts := pm.ListPrompts()
	assert.Len(t, prompts, 2)
	assert.Contains(t, prompts, prompt1)
	assert.Contains(t, prompts, prompt2)

	// Remove
	pm.RemovePrompt("prompt1")
	_, ok = pm.GetPrompt("prompt1")
	assert.False(t, ok)
	assert.Len(t, pm.ListPrompts(), 1)
}

func TestPromptManager_SetMCPServer(t *testing.T) {
	pm := NewPromptManager()
	mockProvider := newMockMCPServerProvider()
	pm.SetMCPServer(mockProvider)
	assert.Equal(t, mockProvider, pm.mcpServer)
}

func TestPromptManager_AddPromptWithMCPServer(t *testing.T) {
	pm := NewPromptManager()
	mockProvider := newMockMCPServerProvider()
	pm.SetMCPServer(mockProvider)

	prompt := &mockPrompt{name: "prompt1", service: "service1"}
	pm.AddPrompt(prompt)

	assert.Len(t, mockProvider.server.addPromptCalls, 1)
	assert.Equal(t, prompt.Prompt(), mockProvider.server.addPromptCalls[0])

	// Verify that the handler calls the prompt's Get method
	req := &mcp.GetPromptRequest{}
	req.Params = &mcp.GetPromptParams{
		Arguments: map[string]string{},
	}
	_, err := mockProvider.server.prompts["prompt1"](context.Background(), req)
	assert.NoError(t, err)
}

func TestPromptManager_RemovePromptWithMCPServer(t *testing.T) {
	pm := NewPromptManager()
	mockProvider := newMockMCPServerProvider()
	pm.SetMCPServer(mockProvider)

	prompt := &mockPrompt{name: "prompt1", service: "service1"}
	pm.AddPrompt(prompt)
	pm.RemovePrompt("prompt1")

	assert.Len(t, mockProvider.server.removePromptsCalls, 1)
	assert.Equal(t, "prompt1", mockProvider.server.removePromptsCalls[0])
}
