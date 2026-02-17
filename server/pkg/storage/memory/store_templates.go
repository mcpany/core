// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// ListServiceTemplates lists all service templates.
//
// Summary: Retrieves all stored service templates.
//
// Parameters:
//   - _ : context.Context. The context (unused).
//
// Returns:
//   - []*configv1.ServiceTemplate: A list of service templates.
//   - error: An error if retrieval fails.
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
// Summary: Retrieves a specific service template.
//
// Parameters:
//   - _ : context.Context. The context (unused).
//   - id: string. The unique identifier of the template.
//
// Returns:
//   - *configv1.ServiceTemplate: The service template.
//   - error: An error if retrieval fails.
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
// Summary: Stores or updates a service template.
//
// Parameters:
//   - _ : context.Context. The context (unused).
//   - template: *configv1.ServiceTemplate. The template to save.
//
// Returns:
//   - error: An error if saving fails.
func (s *Store) SaveServiceTemplate(_ context.Context, template *configv1.ServiceTemplate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceTemplates[template.GetId()] = proto.Clone(template).(*configv1.ServiceTemplate)
	return nil
}

// DeleteServiceTemplate deletes a service template by ID.
//
// Summary: Deletes a service template.
//
// Parameters:
//   - _ : context.Context. The context (unused).
//   - id: string. The ID of the template to delete.
//
// Returns:
//   - error: An error if deletion fails.
func (s *Store) DeleteServiceTemplate(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.serviceTemplates, id)
	return nil
}
