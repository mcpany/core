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

// Manager - Auto-generated documentation.
//
// Summary: Manager manages webhooks.
//
// Fields:
//   - Various fields for Manager.
type Manager struct {
	mu         sync.RWMutex
	webhooks   map[string]*WebhookConfig
	httpClient *http.Client
}

// NewManager - Auto-generated documentation.
//
// Summary: NewManager creates a new Webhook Manager.
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
		webhooks:   make(map[string]*WebhookConfig),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// ListWebhooks - Auto-generated documentation.
//
// Summary: ListWebhooks returns all configured webhooks.
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
func (m *Manager) ListWebhooks() []*WebhookConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*WebhookConfig, 0, len(m.webhooks))
	for _, w := range m.webhooks {
		list = append(list, w)
	}
	return list
}

// AddWebhook - Auto-generated documentation.
//
// Summary: AddWebhook adds or updates a webhook.
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
func (m *Manager) AddWebhook(w *WebhookConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if w.ID == "" {
		w.ID = fmt.Sprintf("wh-%d", time.Now().UnixNano())
	}
	// Ensure active defaults to true if new? Or let caller decide.
	m.webhooks[w.ID] = w
}

// GetWebhook returns a webhook by ID. Summary: Retrieves a webhook by ID. Parameters: - id (string): The webhook ID. Returns: - *WebhookConfig: The webhook configuration. - bool: True if found, false otherwise.
//
// Summary: GetWebhook returns a webhook by ID. Summary: Retrieves a webhook by ID. Parameters: - id (string): The webhook ID. Returns: - *WebhookConfig: The webhook configuration. - bool: True if found, false otherwise.
//
// Parameters:
//   - id (string): The unique identifier used to reference the  resource.
//
// Returns:
//   - (*WebhookConfig): The resulting WebhookConfig object containing the requested data.
//   - (bool): A boolean indicating the success or status of the operation.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *Manager) GetWebhook(id string) (*WebhookConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, ok := m.webhooks[id]
	return w, ok
}

// DeleteWebhook - Auto-generated documentation.
//
// Summary: DeleteWebhook removes a webhook by ID.
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
