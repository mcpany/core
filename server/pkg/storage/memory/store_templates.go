// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// ListServiceTemplates retrieves all service templates.
//
// Summary: Lists all service templates.
//
// Parameters:
//   - ctx: context.Context. The request context.
//
// Returns:
//   - []*configv1.ServiceTemplate: The list of templates.
//   - error: Error if listing fails.
func (s *Store) ListServiceTemplates(_ context.Context) ([]*configv1.ServiceTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.ServiceTemplate, 0, len(s.serviceTemplates))
	for _, t := range s.serviceTemplates {
		list = append(list, proto.Clone(t).(*configv1.ServiceTemplate))
	}
	return list, nil
}

// GetServiceTemplate retrieves a service template by ID.
//
// Summary: Retrieves a service template by ID.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - id: string. The template ID.
//
// Returns:
//   - *configv1.ServiceTemplate: The template.
//   - error: Error if retrieval fails.
func (s *Store) GetServiceTemplate(_ context.Context, id string) (*configv1.ServiceTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if t, ok := s.serviceTemplates[id]; ok {
		return proto.Clone(t).(*configv1.ServiceTemplate), nil
	}
	return nil, nil
}

// SaveServiceTemplate saves a service template.
//
// Summary: Persists a service template.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - template: *configv1.ServiceTemplate. The template to save.
//
// Returns:
//   - error: Error if saving fails.
func (s *Store) SaveServiceTemplate(_ context.Context, template *configv1.ServiceTemplate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceTemplates[template.GetId()] = proto.Clone(template).(*configv1.ServiceTemplate)
	return nil
}

// DeleteServiceTemplate deletes a service template by ID.
//
// Summary: Deletes a service template by ID.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - id: string. The template ID.
//
// Returns:
//   - error: Error if deletion fails.
func (s *Store) DeleteServiceTemplate(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.serviceTemplates, id)
	return nil
}
