// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package webhook provides functionality for managing webhook subscriptions.
package webhook

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Manager manages webhook subscriptions.
type Manager struct {
	mu            sync.RWMutex
	subscriptions map[string]*configv1.WebhookSubscription
}

// NewManager creates a new WebhookManager.
func NewManager() *Manager {
	return &Manager{
		subscriptions: make(map[string]*configv1.WebhookSubscription),
	}
}

// NewManagerWithDefaults creates a new WebhookManager with some default subscriptions for testing.
func NewManagerWithDefaults() *Manager {
	m := NewManager()
	// Seed a default webhook
	wh := configv1.WebhookSubscription_builder{
		Id:            "wh-seed-1",
		Url:           "https://example.com/webhook",
		Events:        []string{"service.registered"},
		Active:        true,
		Secret:        "seeded-secret",
		Status:        "success",
		LastTriggered: "Never",
	}.Build()

	_, _ = m.CreateWebhook(wh)
	return m
}

// ListWebhooks returns all webhook subscriptions.
func (m *Manager) ListWebhooks() []*configv1.WebhookSubscription {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]*configv1.WebhookSubscription, 0, len(m.subscriptions))
	for _, sub := range m.subscriptions {
		res = append(res, sub)
	}
	return res
}

// GetWebhook returns a webhook by ID.
func (m *Manager) GetWebhook(id string) (*configv1.WebhookSubscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sub, ok := m.subscriptions[id]
	if !ok {
		return nil, fmt.Errorf("webhook %s not found", id)
	}
	return sub, nil
}

// CreateWebhook creates a new webhook subscription.
func (m *Manager) CreateWebhook(sub *configv1.WebhookSubscription) (*configv1.WebhookSubscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sub.GetId() == "" {
		sub.SetId("wh_" + generateRandomID())
	}
	if sub.GetSecret() == "" {
		sub.SetSecret("sec_" + generateRandomID())
	}
	sub.SetStatus("pending")
	sub.SetLastTriggered("Never")

	m.subscriptions[sub.GetId()] = sub
	return sub, nil
}

// DeleteWebhook deletes a webhook subscription.
func (m *Manager) DeleteWebhook(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.subscriptions[id]; !ok {
		return fmt.Errorf("webhook %s not found", id)
	}
	delete(m.subscriptions, id)
	return nil
}

func generateRandomID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
