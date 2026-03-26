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

// ManagerInterface managerInterface represents a manager interface.
//
// Summary: ManagerInterface represents a manager interface.
type ManagerInterface interface {
	// AddPrompt registers a new prompt.
	//
	// Parameters: - None.
	//   - prompt: Prompt. The prompt definition to add.
	AddPrompt(prompt Prompt)

	// UpdatePrompt updates an existing prompt.
	//
	// Parameters: - None.
	//   - prompt: Prompt. The prompt with updated information.
	UpdatePrompt(prompt Prompt)

	// GetPrompt retrieves a prompt by its name.
	//
	// Parameters: - None.
	//   - name: string. The unique name of the prompt.
	//
	// Returns: - None.
	//   - Prompt: The prompt instance.
	//   - bool: True if the prompt was found, false otherwise.
	GetPrompt(name string) (Prompt, bool)

	// ListPrompts returns all registered prompts.
	//
	// Returns: - None.
	//   - []Prompt: A slice of all registered prompts.
	ListPrompts() []Prompt

	// ClearPromptsForService removes all prompts associated with a specific service.
	//
	// Parameters: - None.
	//   - serviceID: string. The unique identifier of the service.
	ClearPromptsForService(serviceID string)

	// SetMCPServer sets the MCP server provider.
	//
	// Parameters: - None.
	//   - mcpServer: MCPServerProvider. The provider interface.
	SetMCPServer(mcpServer MCPServerProvider)
}

// Manager manager represents a manager.
//
// Summary: Manager represents a manager.
type Manager struct {
	prompts       *xsync.Map[string, Prompt]
	mcpServer     MCPServerProvider
	mu            sync.RWMutex
	cachedPrompts []Prompt
}

// NewManager creates and returns a new, empty Manager.
//
// Returns: - None.
//   - *Manager: A pointer to the newly created Manager.
//
// Summary: Initializes NewManager operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func NewManager() *Manager {
	return &Manager{
		prompts: xsync.NewMap[string, Prompt](),
	}
}

// SetMCPServer provides the Manager with a reference to the MCP server.
//
// Parameters: - None.
//   - mcpServer: MCPServerProvider. The MCP server provider.
//
// Summary: Updates SetMCPServer operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (pm *Manager) SetMCPServer(mcpServer MCPServerProvider) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.mcpServer = mcpServer
}

// AddPrompt registers a new prompt with the manager.
//
// If a prompt with the same name already exists, it will be overwritten, and a warning
// will be logged.
//
// Parameters: - None.
//   - prompt: Prompt. The prompt to add.
//
// Side Effects: - None.
//   - Updates the internal prompt registry.
//   - Invalidates the list cache.
//
// Summary: Executes AddPrompt operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
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

// UpdatePrompt updates an existing prompt in the manager.
//
// If the prompt does not exist, it will be added.
//
// Parameters: - None.
//   - prompt: Prompt. The prompt definition to update.
//
// Side Effects: - None.
//   - Updates the internal prompt registry.
//   - Invalidates the list cache.
//
// Summary: Executes UpdatePrompt operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (pm *Manager) UpdatePrompt(prompt Prompt) {
	pm.prompts.Store(prompt.Prompt().Name, prompt)
	pm.mu.Lock()
	pm.cachedPrompts = nil
	pm.mu.Unlock()
}

// GetPrompt retrieves a prompt from the manager by its name.
//
// Parameters: - None.
//   - name: string. The name of the prompt.
//
// Returns: - None.
//   - Prompt: The prompt instance.
//   - bool: True if found, false otherwise.
//
// Summary: Retrieves GetPrompt operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (pm *Manager) GetPrompt(name string) (Prompt, bool) {
	prompt, ok := pm.prompts.Load(name)
	return prompt, ok
}

// ListPrompts returns a slice containing all the prompts currently registered.
//
// It uses a read-through cache to improve performance.
//
// Returns: - None.
//   - []Prompt: A slice of currently registered prompts.
//
// Summary: Executes ListPrompts operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
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
// Parameters: - None.
//   - serviceID: string. The unique identifier of the service.
//
// Side Effects: - None.
//   - Removes matching prompts from the registry.
//   - Invalidates the list cache.
//
// Summary: Executes ClearPromptsForService operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
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
