package prompt

import (
	"sync"
)

// PromptManagerInterface defines the interface for managing prompts.
type PromptManagerInterface interface {
	AddPrompt(prompt *Prompt)
	ListPrompts() []*Prompt
	GetPrompt(name string) (*Prompt, bool)
	ClearPromptsForService(serviceID string)
}

// PromptManager is a concrete implementation of PromptManagerInterface.
type PromptManager struct {
	mu      sync.RWMutex
	prompts map[string]*Prompt
}

// NewPromptManager creates a new PromptManager.
func NewPromptManager() *PromptManager {
	return &PromptManager{
		prompts: make(map[string]*Prompt),
	}
}

// AddPrompt adds a new prompt to the manager.
func (m *PromptManager) AddPrompt(prompt *Prompt) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.prompts[prompt.Name] = prompt
}

// ListPrompts returns a list of all registered prompts.
func (m *PromptManager) ListPrompts() []*Prompt {
	m.mu.RLock()
	defer m.mu.RUnlock()

	prompts := make([]*Prompt, 0, len(m.prompts))
	for _, p := range m.prompts {
		prompts = append(prompts, p)
	}
	return prompts
}

// GetPrompt returns a prompt by name.
func (m *PromptManager) GetPrompt(name string) (*Prompt, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.prompts[name]
	return p, ok
}

// ClearPromptsForService removes all prompts associated with a service.
func (m *PromptManager) ClearPromptsForService(serviceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, p := range m.prompts {
		if p.ServiceID == serviceID {
			delete(m.prompts, name)
		}
	}
}
