// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookAuditStore sends audit logs to a configured webhook URL.
type WebhookAuditStore struct {
	webhookURL string
	headers    map[string]string
	client     *http.Client
}

// NewWebhookAuditStore creates a new WebhookAuditStore.
func NewWebhookAuditStore(webhookURL string, headers map[string]string) *WebhookAuditStore {
	return &WebhookAuditStore{
		webhookURL: webhookURL,
		headers:    headers,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

// Write writes an audit entry to the webhook.
func (s *WebhookAuditStore) Write(ctx context.Context, entry AuditEntry) error {
	payload, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range s.headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status: %s", resp.Status)
	}

	return nil
}

// Close is a no-op for WebhookAuditStore.
func (s *WebhookAuditStore) Close() error {
	return nil
}
