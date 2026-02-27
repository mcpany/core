// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock prompt for testing
type mockPrompt struct {
	name      string
	serviceID string
}

func (m *mockPrompt) Prompt() *mcp.Prompt {
	return &mcp.Prompt{Name: m.name}
}

func (m *mockPrompt) Service() string {
	return m.serviceID
}

func (m *mockPrompt) Definition() *configv1.PromptDefinition {
	return nil
}

func (m *mockPrompt) Get(_ context.Context, _ json.RawMessage) (*mcp.GetPromptResult, error) {
	return nil, nil
}

func TestManager_AddAndList(t *testing.T) {
	m := NewManager()

	p1 := &mockPrompt{name: "p1", serviceID: "s1"}
	m.AddPrompt(p1)

	list := m.ListPrompts()
	require.Len(t, list, 1)
	assert.Equal(t, "p1", list[0].Prompt().Name)

	p2 := &mockPrompt{name: "p2", serviceID: "s2"}
	m.AddPrompt(p2)

	list = m.ListPrompts()
	require.Len(t, list, 2)
}

func TestManager_Update(t *testing.T) {
	m := NewManager()
	p1 := &mockPrompt{name: "p1", serviceID: "s1"}
	m.AddPrompt(p1)

	// Update with same name but different service
	p1Update := &mockPrompt{name: "p1", serviceID: "s2"}
	m.UpdatePrompt(p1Update)

	p, ok := m.GetPrompt("p1")
	require.True(t, ok)
	assert.Equal(t, "s2", p.Service())
}

func TestManager_ClearPromptsForService(t *testing.T) {
	m := NewManager()
	m.AddPrompt(&mockPrompt{name: "p1", serviceID: "s1"})
	m.AddPrompt(&mockPrompt{name: "p2", serviceID: "s1"})
	m.AddPrompt(&mockPrompt{name: "p3", serviceID: "s2"})

	m.ClearPromptsForService("s1")

	list := m.ListPrompts()
	require.Len(t, list, 1)
	assert.Equal(t, "p3", list[0].Prompt().Name)
}

func TestManager_CacheInvalidation(t *testing.T) {
	m := NewManager()
	p1 := &mockPrompt{name: "p1", serviceID: "s1"}

	// Add -> Populate Cache
	m.AddPrompt(p1)
	list1 := m.ListPrompts()
	assert.Len(t, list1, 1)

	// Verify internal state (whitebox testing)
	m.mu.RLock()
	assert.NotNil(t, m.cachedPrompts)
	m.mu.RUnlock()

	// Update -> Invalidate Cache
	p2 := &mockPrompt{name: "p2", serviceID: "s1"}
	m.AddPrompt(p2)

	m.mu.RLock()
	assert.Nil(t, m.cachedPrompts)
	m.mu.RUnlock()

	// List -> Repopulate Cache
	list2 := m.ListPrompts()
	assert.Len(t, list2, 2)
}

func TestManager_ConcurrentAccess(t *testing.T) {
	m := NewManager()
	concurrency := 50
	var wg sync.WaitGroup

	// Writer goroutines
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("p-%d", id)
			m.AddPrompt(&mockPrompt{name: name, serviceID: "s1"})
			// Occasionally update or clear
			if id%10 == 0 {
				m.UpdatePrompt(&mockPrompt{name: name, serviceID: "s2"})
			}
		}(i)
	}

	// Reader goroutines
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Just calling ListPrompts to trigger race detector on cache access
			_ = m.ListPrompts()
		}()
	}

	// Clear goroutines
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.ClearPromptsForService("s1")
		}()
	}

	wg.Wait()

	// Final consistency check
	// Should not panic or have races
	_ = m.ListPrompts()
}
