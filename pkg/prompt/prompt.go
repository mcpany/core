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

package prompt

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ErrPromptNotFound is returned when a requested prompt cannot be found.
var ErrPromptNotFound = errors.New("prompt not found")

// Prompt defines the interface for a prompt that can be managed by the
// PromptManager. Each implementation of a prompt is responsible for providing its
// metadata and handling its execution.
type Prompt interface {
	// Prompt returns the MCP representation of the prompt, which includes its
	// metadata.
	Prompt() *mcp.Prompt
	// Service returns the ID of the service that provides this prompt.
	Service() string
	// Get executes the prompt with the given arguments and returns the result.
	Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error)
}

// MCPServerProvider defines an interface for components that can provide an
// instance of an *mcp.Server. This is used to decouple the PromptManager from the
// concrete server implementation.
type MCPServerProvider interface {
	Server() *mcp.Server
}

// PromptManagerInterface defines the interface for managing a collection of
// prompts. It provides methods for adding, removing, retrieving, and listing
// prompts.
type PromptManagerInterface interface {
	// GetPrompt retrieves a prompt by its name.
	GetPrompt(name string) (Prompt, bool)
	// AddPrompt adds a new prompt to the manager.
	AddPrompt(prompt Prompt)
	// RemovePrompt removes a prompt from the manager by its name.
	RemovePrompt(name string)
	// ListPrompts returns a slice of all prompts currently in the manager.
	ListPrompts() []Prompt
	// SetMCPServer provides the PromptManager with a reference to the MCP server.
	SetMCPServer(mcpServer MCPServerProvider)
}

// PromptManager is a thread-safe implementation of the PromptManagerInterface.
// It uses a map to store prompts and a mutex to protect concurrent access.
type PromptManager struct {
	mu        sync.RWMutex
	prompts   map[string]Prompt
	mcpServer MCPServerProvider
}

// NewPromptManager creates and returns a new, empty PromptManager.
func NewPromptManager() *PromptManager {
	return &PromptManager{
		prompts: make(map[string]Prompt),
	}
}

// GetPrompt retrieves a prompt from the manager by its name.
//
// name is the name of the prompt to retrieve.
// It returns the prompt and a boolean indicating whether the prompt was found.
func (pm *PromptManager) GetPrompt(name string) (Prompt, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	prompt, ok := pm.prompts[name]
	return prompt, ok
}

// AddPrompt adds a new prompt to the manager. If a prompt with the same name
// already exists, it will be overwritten.
//
// prompt is the prompt to be added.
func (pm *PromptManager) AddPrompt(prompt Prompt) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.prompts[prompt.Prompt().Name] = prompt
	if pm.mcpServer != nil {
		pm.mcpServer.Server().AddPrompt(prompt.Prompt(), func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			argsBytes, err := json.Marshal(req.Params.Arguments)
			if err != nil {
				return nil, err
			}
			return prompt.Get(ctx, argsBytes)
		})
	}
}

// RemovePrompt removes a prompt from the manager by its name.
//
// name is the name of the prompt to be removed.
func (pm *PromptManager) RemovePrompt(name string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if _, ok := pm.prompts[name]; ok {
		delete(pm.prompts, name)
		if pm.mcpServer != nil {
			pm.mcpServer.Server().RemovePrompts(name)
		}
	}
}

// ListPrompts returns a slice containing all the prompts currently registered in
// the manager.
func (pm *PromptManager) ListPrompts() []Prompt {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	prompts := make([]Prompt, 0, len(pm.prompts))
	for _, prompt := range pm.prompts {
		prompts = append(prompts, prompt)
	}
	return prompts
}

// SetMCPServer provides the PromptManager with a reference to the MCP server.
// This is necessary for registering prompt handlers with the server.
func (pm *PromptManager) SetMCPServer(mcpServer MCPServerProvider) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.mcpServer = mcpServer
}
