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

// ManagerInterface - Auto-generated documentation.
//
// Summary: ManagerInterface defines the interface for a prompt manager.
//
// Methods:
//   - Various methods for ManagerInterface.
type ManagerInterface interface {
	// AddPrompt registers a new prompt.
	//
	// Parameters:
	//   - prompt: Prompt. The prompt definition to add.
	AddPrompt(prompt Prompt)

	// UpdatePrompt updates an existing prompt.
	//
	// Parameters:
	//   - prompt: Prompt. The prompt with updated information.
	UpdatePrompt(prompt Prompt)

	// GetPrompt retrieves a prompt by its name.
	//
	// Parameters:
	//   - name: string. The unique name of the prompt.
	//
	// Returns:
	//   - Prompt: The prompt instance.
	//   - bool: True if the prompt was found, false otherwise.
	GetPrompt(name string) (Prompt, bool)

	// ListPrompts returns all registered prompts.
	//
	// Returns:
	//   - []Prompt: A slice of all registered prompts.
	ListPrompts() []Prompt

	// ClearPromptsForService removes all prompts associated with a specific service.
	//
	// Parameters:
	//   - serviceID: string. The unique identifier of the service.
	ClearPromptsForService(serviceID string)

	// SetMCPServer sets the MCP server provider.
	//
	// Parameters:
	//   - mcpServer: MCPServerProvider. The provider interface.
	SetMCPServer(mcpServer MCPServerProvider)
}

// Manager - Auto-generated documentation.
//
// Summary: Manager is a thread-safe manager for registering and retrieving prompts.
//
// Fields:
//   - Various fields for Manager.
type Manager struct {
	prompts       *xsync.Map[string, Prompt]
	mcpServer     MCPServerProvider
	mu            sync.RWMutex
	cachedPrompts []Prompt
}

// NewManager - Auto-generated documentation.
//
// Summary: NewManager creates and returns a new, empty Manager.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func NewManager() *Manager {
	return &Manager{
		prompts: xsync.NewMap[string, Prompt](),
	}
}

// SetMCPServer - Auto-generated documentation.
//
// Summary: SetMCPServer provides the Manager with a reference to the MCP server.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (pm *Manager) SetMCPServer(mcpServer MCPServerProvider) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.mcpServer = mcpServer
}

// AddPrompt - Auto-generated documentation.
//
// Summary: AddPrompt registers a new prompt with the manager.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
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

// UpdatePrompt - Auto-generated documentation.
//
// Summary: UpdatePrompt updates an existing prompt in the manager.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (pm *Manager) UpdatePrompt(prompt Prompt) {
	pm.prompts.Store(prompt.Prompt().Name, prompt)
	pm.mu.Lock()
	pm.cachedPrompts = nil
	pm.mu.Unlock()
}

// GetPrompt retrieves a prompt from the manager by its name. Parameters: - name: string. The name of the prompt. Returns: - Prompt: The prompt instance. - bool: True if found, false otherwise.
//
// Summary: GetPrompt retrieves a prompt from the manager by its name. Parameters: - name: string. The name of the prompt. Returns: - Prompt: The prompt instance. - bool: True if found, false otherwise.
//
// Parameters:
//   - name (string): The name parameter used in the operation.
//
// Returns:
//   - (Prompt): The resulting Prompt object containing the requested data.
//   - (bool): A boolean indicating the success or status of the operation.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (pm *Manager) GetPrompt(name string) (Prompt, bool) {
	prompt, ok := pm.prompts.Load(name)
	return prompt, ok
}

// ListPrompts - Auto-generated documentation.
//
// Summary: ListPrompts returns a slice containing all the prompts currently registered.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
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

// ClearPromptsForService - Auto-generated documentation.
//
// Summary: ClearPromptsForService removes all prompts associated with a given service.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
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
