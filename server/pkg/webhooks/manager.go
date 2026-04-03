// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// WebhookConfig represents the public WebhookConfig entity.
//
// Summary: Defines the structured data model representing a config.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type WebhookConfig struct {
	ID            string    `json:"id"`
	URL           string    `json:"url"`
	Events        []string  `json:"events"`
	Active        bool      `json:"active"`
	LastTriggered time.Time `json:"last_triggered,omitempty"`
	Status        string    `json:"status,omitempty"` // success, failure, pending
}

// Manager represents the public Manager entity.
//
// Summary: Coordinates operations and orchestrates lifecycle events for the  components.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Manager struct {
	mu         sync.RWMutex
	webhooks   map[string]*WebhookConfig
	httpClient *http.Client
}

// NewManager serves as a public interface for interacting with NewManager.
//
// Summary: Constructs and returns an initialized manager ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewManager() *Manager {
	return &Manager{
		webhooks:   make(map[string]*WebhookConfig),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// ListWebhooks serves as a public interface for interacting with ListWebhooks.
//
// Summary: List the webhooks appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *Manager) ListWebhooks() []*WebhookConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*WebhookConfig, 0, len(m.webhooks))
	for _, w := range m.webhooks {
		list = append(list, w)
	}
	return list
}

// AddWebhook serves as a public interface for interacting with AddWebhook.
//
// Summary: Add the webhook appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *Manager) AddWebhook(w *WebhookConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if w.ID == "" {
		w.ID = fmt.Sprintf("wh-%d", time.Now().UnixNano())
	}
	// Ensure active defaults to true if new? Or let caller decide.
	m.webhooks[w.ID] = w
}

// GetWebhook serves as a public interface for interacting with GetWebhook.
//
// Summary: Fetches and returns the underlying webhook from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *Manager) GetWebhook(id string) (*WebhookConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, ok := m.webhooks[id]
	return w, ok
}

// DeleteWebhook serves as a public interface for interacting with DeleteWebhook.
//
// Summary: Delete the webhook appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *Manager) DeleteWebhook(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.webhooks, id)
}

// TestWebhook serves as a public interface for interacting with TestWebhook.
//
// Summary: Test the webhook appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *Manager) TestWebhook(ctx context.Context, id string) error {
	w, ok := m.GetWebhook(id)
	if !ok {
		return fmt.Errorf("webhook not found")
	}

	// Mock payload
	payload := []byte(`{"event": "test", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`)
	req, err := http.NewRequestWithContext(ctx, "POST", w.URL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		m.updateStatus(id, "failure")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		m.updateStatus(id, "success")
		return nil
	}

	m.updateStatus(id, "failure")
	return fmt.Errorf("status code: %d", resp.StatusCode)
}

func (m *Manager) updateStatus(id, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if w, ok := m.webhooks[id]; ok {
		w.Status = status
		w.LastTriggered = time.Now()
	}
}
