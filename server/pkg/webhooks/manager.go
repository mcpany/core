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

// WebhookConfig represents a configured webhook.
//
// Summary: Configuration for a webhook.
type WebhookConfig struct {
	ID            string    `json:"id"`
	URL           string    `json:"url"`
	Events        []string  `json:"events"`
	Active        bool      `json:"active"`
	LastTriggered time.Time `json:"last_triggered,omitempty"`
	Status        string    `json:"status,omitempty"` // success, failure, pending
}

// Manager manages webhooks.
//
// Summary: Manager for handling webhook lifecycle and execution.
type Manager struct {
	mu         sync.RWMutex
	webhooks   map[string]*WebhookConfig
	httpClient *http.Client
}

// NewManager creates a new Webhook Manager.
//
// Summary: Initializes a new Webhook Manager.
//
// Returns:
//   - *Manager: A pointer to the newly created Manager.
func NewManager() *Manager {
	return &Manager{
		webhooks:   make(map[string]*WebhookConfig),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// ListWebhooks returns all configured webhooks.
//
// Summary: Retrieves a list of all configured webhooks.
//
// Returns:
//   - []*WebhookConfig: A slice of all configured webhooks.
func (m *Manager) ListWebhooks() []*WebhookConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*WebhookConfig, 0, len(m.webhooks))
	for _, w := range m.webhooks {
		list = append(list, w)
	}
	return list
}

// AddWebhook adds or updates a webhook.
//
// Summary: Adds a new webhook or updates an existing one.
//
// Parameters:
//   - w: *WebhookConfig. The webhook configuration to add or update. If ID is empty, one will be generated.
//
// Returns:
//   - None.
func (m *Manager) AddWebhook(w *WebhookConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if w.ID == "" {
		w.ID = fmt.Sprintf("wh-%d", time.Now().UnixNano())
	}
	// Ensure active defaults to true if new? Or let caller decide.
	m.webhooks[w.ID] = w
}

// GetWebhook returns a webhook by ID.
//
// Summary: Retrieves a specific webhook by its ID.
//
// Parameters:
//   - id: string. The unique identifier of the webhook.
//
// Returns:
//   - *WebhookConfig: The requested webhook if found.
//   - bool: True if the webhook exists, false otherwise.
func (m *Manager) GetWebhook(id string) (*WebhookConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, ok := m.webhooks[id]
	return w, ok
}

// DeleteWebhook removes a webhook by ID.
//
// Summary: Deletes a webhook configuration.
//
// Parameters:
//   - id: string. The unique identifier of the webhook to delete.
//
// Returns:
//   - None.
func (m *Manager) DeleteWebhook(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.webhooks, id)
}

// TestWebhook sends a test payload to the webhook URL.
//
// Summary: Triggers a test event for the specified webhook.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - id: string. The unique identifier of the webhook to test.
//
// Returns:
//   - error: Returns nil on success, or an error if the webhook is not found or the request fails.
//
// Throws/Errors:
//   - Returns error if "webhook not found".
//   - Returns error if the HTTP request fails or returns a non-2xx status code.
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
