// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"
	"fmt"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// CreateSystemWebhook creates a new system webhook.
func (s *Store) CreateSystemWebhook(_ context.Context, webhook *configv1.SystemWebhook) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.systemWebhooks[webhook.GetId()]; ok {
		return fmt.Errorf("system webhook already exists")
	}
	s.systemWebhooks[webhook.GetId()] = proto.Clone(webhook).(*configv1.SystemWebhook)
	return nil
}

// ListSystemWebhooks retrieves all system webhooks.
func (s *Store) ListSystemWebhooks(_ context.Context) ([]*configv1.SystemWebhook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.SystemWebhook, 0, len(s.systemWebhooks))
	for _, w := range s.systemWebhooks {
		list = append(list, proto.Clone(w).(*configv1.SystemWebhook))
	}
	return list, nil
}

// GetSystemWebhook retrieves a system webhook by ID.
func (s *Store) GetSystemWebhook(_ context.Context, id string) (*configv1.SystemWebhook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if w, ok := s.systemWebhooks[id]; ok {
		return proto.Clone(w).(*configv1.SystemWebhook), nil
	}
	return nil, nil
}

// UpdateSystemWebhook updates an existing system webhook.
func (s *Store) UpdateSystemWebhook(_ context.Context, webhook *configv1.SystemWebhook) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.systemWebhooks[webhook.GetId()]; !ok {
		return fmt.Errorf("system webhook not found")
	}
	s.systemWebhooks[webhook.GetId()] = proto.Clone(webhook).(*configv1.SystemWebhook)
	return nil
}

// DeleteSystemWebhook deletes a system webhook by ID.
func (s *Store) DeleteSystemWebhook(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.systemWebhooks, id)
	return nil
}
