// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package prompt provides prompt management functionality.
package prompt

import (
	"fmt"
	"sync"

	"github.com/mcpany/core/server/pkg/logging"
	xsync "github.com/puzpuzpuz/xsync/v4"
)

// ManagerInterface defines the interface for a prompt manager.
type ManagerInterface interface {
	// AddPrompt registers a new prompt.
	//
	// prompt is the prompt.
	AddPrompt(prompt Prompt)
	// UpdatePrompt updates an existing prompt.
	//
	// prompt is the prompt.
	UpdatePrompt(prompt Prompt)
	// GetPrompt retrieves a prompt by name.
	//
	// name is the name of the resource.
	//
	// Returns the result.
	// Returns true if successful.
	GetPrompt(name string) (Prompt, bool)
	// ListPrompts returns all registered prompts.
	//
	// Returns the result.
	ListPrompts() []Prompt
	// ClearPromptsForService removes all prompts associated with a service.
	//
	// serviceID is the serviceID.
	ClearPromptsForService(serviceID string)
	// SetMCPServer sets the MCP server provider.
	//
	// mcpServer is the mcpServer.
	SetMCPServer(mcpServer MCPServerProvider)
}

// Manager is a thread-safe manager for registering and retrieving prompts.
type Manager struct {
	prompts       *xsync.Map[string, Prompt]
	mcpServer     MCPServerProvider
	mu            sync.RWMutex
	cachedPrompts []Prompt
}

// NewManager creates and returns a new, empty Manager.
//
// Returns the result.
func NewManager() *Manager {
	return &Manager{
		prompts: xsync.NewMap[string, Prompt](),
	}
}

// SetMCPServer provides the Manager with a reference to the MCP server.
//
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
	pm.mu.Lock()
	pm.cachedPrompts = nil
	pm.mu.Unlock()
}

// UpdatePrompt updates an existing prompt in the manager. If the prompt does not
// exist, it will be added.
func (pm *Manager) UpdatePrompt(prompt Prompt) {
	pm.prompts.Store(prompt.Prompt().Name, prompt)
	pm.mu.Lock()
	pm.cachedPrompts = nil
	pm.mu.Unlock()
}

// GetPrompt retrieves a prompt from the manager by its name.
//
// name is the name of the resource.
//
// Returns the result.
// Returns true if successful.
func (pm *Manager) GetPrompt(name string) (Prompt, bool) {
	prompt, ok := pm.prompts.Load(name)
	return prompt, ok
}

// ListPrompts returns a slice containing all the prompts currently registered.
//
// Returns the result.
func (pm *Manager) ListPrompts() []Prompt {
	// ⚡ Bolt: Use a read-through cache to avoid repeated map iteration and slice allocation.
	// The cache is invalidated on any write operation (Add/Update/Clear).
	// We use double-checked locking to safely upgrade from RLock to Lock.
	pm.mu.RLock()
	if pm.cachedPrompts != nil {
		// Return a copy to ensure thread safety
		result := make([]Prompt, len(pm.cachedPrompts))
		copy(result, pm.cachedPrompts)
		pm.mu.RUnlock()
		return result
	}
	pm.mu.RUnlock()

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Double-check after acquiring the write lock
	if pm.cachedPrompts != nil {
		// Return a copy to ensure thread safety
		result := make([]Prompt, len(pm.cachedPrompts))
		copy(result, pm.cachedPrompts)
		return result
	}

	prompts := make([]Prompt, 0)
	pm.prompts.Range(func(_ string, value Prompt) bool {
		prompts = append(prompts, value)
		return true
	})
	pm.cachedPrompts = prompts

	// Return a copy to ensure thread safety
	result := make([]Prompt, len(prompts))
	copy(result, prompts)
	return result
}

// ClearPromptsForService removes all prompts associated with a given service.
//
// serviceID is the serviceID.
func (pm *Manager) ClearPromptsForService(serviceID string) {
	changed := false
	pm.prompts.Range(func(key string, value Prompt) bool {
		if value.Service() == serviceID {
			pm.prompts.Delete(key)
			changed = true
		}
		return true
	})

	if changed {
		pm.mu.Lock()
		pm.cachedPrompts = nil
		pm.mu.Unlock()
	}
}
