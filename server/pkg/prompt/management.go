// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package prompt provides prompt management functionality.
package prompt

import (
	"fmt"
	"sync"

	"github.com/mcpany/core/pkg/logging"
	xsync "github.com/puzpuzpuz/xsync/v4"
)

// ManagerInterface defines the interface for a prompt manager.
type ManagerInterface interface {
	// AddPrompt registers a new prompt.
	AddPrompt(prompt Prompt)
	// UpdatePrompt updates an existing prompt.
	UpdatePrompt(prompt Prompt)
	// GetPrompt retrieves a prompt by name.
	GetPrompt(name string) (Prompt, bool)
	// ListPrompts returns all registered prompts.
	ListPrompts() []Prompt
	// ClearPromptsForService removes all prompts associated with a service.
	ClearPromptsForService(serviceID string)
	// SetMCPServer sets the MCP server provider.
	SetMCPServer(mcpServer MCPServerProvider)
}

// Manager is a thread-safe manager for registering and retrieving prompts.
type Manager struct {
	prompts   *xsync.Map[string, Prompt]
	mcpServer MCPServerProvider
	mu        sync.RWMutex
}

// NewManager creates and returns a new, empty Manager.
func NewManager() *Manager {
	return &Manager{
		prompts: xsync.NewMap[string, Prompt](),
	}
}

// SetMCPServer provides the Manager with a reference to the MCP server.
// mcpServer is the mcpServer.
func (pm *Manager) SetMCPServer(mcpServer MCPServerProvider) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.mcpServer = mcpServer
}

// AddPrompt registers a new prompt with the manager. If a prompt with the same
// name already exists, it will be overwritten, and a warning will be logged.
func (pm *Manager) AddPrompt(prompt Prompt) {
	promptName := prompt.Prompt().Name
	if existingPrompt, loaded := pm.prompts.LoadAndStore(promptName, prompt); loaded {
		logging.GetLogger().Warn(fmt.Sprintf("Prompt with the same name already exists. Overwriting. promptName=%s, newPromptService=%s, existingPromptService=%s",
			promptName,
			prompt.Service(),
			existingPrompt.Service(),
		))
	}
}

// UpdatePrompt updates an existing prompt in the manager. If the prompt does not
// exist, it will be added.
func (pm *Manager) UpdatePrompt(prompt Prompt) {
	pm.prompts.Store(prompt.Prompt().Name, prompt)
}

// GetPrompt retrieves a prompt from the manager by its name.
// Returns the result, the result.
func (pm *Manager) GetPrompt(name string) (Prompt, bool) {
	prompt, ok := pm.prompts.Load(name)
	return prompt, ok
}

// ListPrompts returns a slice containing all the prompts currently registered.
func (pm *Manager) ListPrompts() []Prompt {
	var prompts []Prompt
	pm.prompts.Range(func(_ string, value Prompt) bool {
		prompts = append(prompts, value)
		return true
	})
	return prompts
}

// ClearPromptsForService removes all prompts associated with a given service.
// serviceID is the serviceID.
func (pm *Manager) ClearPromptsForService(serviceID string) {
	pm.prompts.Range(func(key string, value Prompt) bool {
		if value.Service() == serviceID {
			pm.prompts.Delete(key)
		}
		return true
	})
}
