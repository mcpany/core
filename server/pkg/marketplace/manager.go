// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package marketplace

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Manager handles marketplace subscriptions and collections.
type Manager interface {
	ListSubscriptions() []Subscription
	AddSubscription(sub Subscription) (Subscription, error)
	GetSubscription(id string) (Subscription, bool)
	UpdateSubscription(id string, sub Subscription) (Subscription, error)
	DeleteSubscription(id string) error
	SyncSubscription(ctx context.Context, id string) error
}

type memoryManager struct {
	mu            sync.RWMutex
	subscriptions map[string]Subscription
}

// NewManager creates a new instance of the marketplace Manager.
func NewManager() Manager {
	m := &memoryManager{
		subscriptions: make(map[string]Subscription),
	}
	// Seed default popular services
	defaultSub := Subscription{
		ID:          "popular",
		Name:        DefaultPopularServices.Name,
		Description: DefaultPopularServices.Description,
		SourceURL:   "internal://popular",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		LastSynced:  time.Now(),
		Services:    DefaultPopularServices.Services,
	}
	m.subscriptions[defaultSub.ID] = defaultSub
	return m
}

func (m *memoryManager) ListSubscriptions() []Subscription {
	m.mu.RLock()
	defer m.mu.RUnlock()
	subs := make([]Subscription, 0, len(m.subscriptions))
	for _, s := range m.subscriptions {
		subs = append(subs, s)
	}
	return subs
}

func (m *memoryManager) AddSubscription(sub Subscription) (Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if sub.ID == "" {
		sub.ID = uuid.New().String()
	}
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()
	m.subscriptions[sub.ID] = sub
	return sub, nil
}

func (m *memoryManager) GetSubscription(id string) (Subscription, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sub, ok := m.subscriptions[id]
	return sub, ok
}

func (m *memoryManager) UpdateSubscription(id string, sub Subscription) (Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.subscriptions[id]
	if !ok {
		return Subscription{}, fmt.Errorf("subscription not found: %s", id)
	}
	// updating fields
	existing.Name = sub.Name
	existing.Description = sub.Description
	existing.SourceURL = sub.SourceURL
	existing.IsActive = sub.IsActive
	existing.UpdatedAt = time.Now()
	m.subscriptions[id] = existing
	return existing, nil
}

func (m *memoryManager) DeleteSubscription(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.subscriptions, id)
	return nil
}

func (m *memoryManager) SyncSubscription(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.subscriptions[id]
	if !ok {
		return fmt.Errorf("subscription not found: %s", id)
	}

	if existing.SourceURL == "internal://popular" {
		existing.Services = DefaultPopularServices.Services
	}
	// TODO: Implement external fetching

	existing.LastSynced = time.Now()
	m.subscriptions[id] = existing
	return nil
}
