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
// Summary: Webhook configuration definition.
//
// Fields:
//   - ID (string): Unique identifier for the webhook.
//   - URL (string): The destination URL.
//   - Events ([]string): List of events to subscribe to.
//   - Active (bool): Whether the webhook is enabled.
//   - LastTriggered (time.Time): Timestamp of the last execution.
//   - Status (string): Status of the last execution (success, failure, pending).
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
// Summary: Webhook lifecycle manager.
type Manager struct {
	mu         sync.RWMutex
	webhooks   map[string]*WebhookConfig
	httpClient *http.Client
}

// NewManager creates a new Webhook Manager.
//
// Summary: Creates a new Manager.
//
// Returns:
//   - *Manager: A pointer to the newly created Manager.
//
// Side Effects:
//   - Initializes internal maps and HTTP client.
func NewManager() *Manager {
	return &Manager{
		webhooks:   make(map[string]*WebhookConfig),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// ListWebhooks returns all configured webhooks.
//
// Summary: Lists all webhooks.
//
// Returns:
//   - []*WebhookConfig: A list of webhook configurations.
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
// Summary: Adds or updates a webhook.
//
// Parameters:
//   - w (*WebhookConfig): The webhook configuration to add.
//
// Side Effects:
//   - Updates the internal webhook map.
//   - Generates an ID if one is not provided.
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
// Summary: Retrieves a webhook by ID.
//
// Parameters:
//   - id (string): The webhook ID.
//
// Returns:
//   - *WebhookConfig: The webhook configuration.
//   - bool: True if found, false otherwise.
func (m *Manager) GetWebhook(id string) (*WebhookConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, ok := m.webhooks[id]
	return w, ok
}

// DeleteWebhook removes a webhook by ID.
//
// Summary: Deletes a webhook.
//
// Parameters:
//   - id (string): The webhook ID to delete.
//
// Side Effects:
//   - Removes the webhook from the internal map.
func (m *Manager) DeleteWebhook(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.webhooks, id)
}

// TestWebhook sends a test payload to the webhook URL.
//
// Summary: Tests a webhook.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - id (string): The webhook ID to test.
//
// Returns:
//   - error: An error if the test fails or the webhook is not found.
//
// Errors:
//   - Returns error if webhook not found.
//   - Returns error if HTTP request fails or returns non-2xx status.
//
// Side Effects:
//   - Sends an HTTP POST request to the webhook URL.
//   - Updates the webhook status.
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
