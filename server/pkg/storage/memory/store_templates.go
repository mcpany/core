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
// Summary: Retrieves all service templates from memory.
//
// Parameters:
//   - ctx: context.Context. Context for the operation (unused in memory store).
//
// Returns:
//   - []*configv1.ServiceTemplate: A slice of all service templates.
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Retrieves a service template from memory.
//
// Parameters:
//   - ctx: context.Context. Context for the operation (unused in memory store).
//   - id: string. The unique identifier of the service template.
//
// Returns:
//   - *configv1.ServiceTemplate: The requested service template if found, or nil if not found.
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Persists a service template to memory.
//
// Parameters:
//   - ctx: context.Context. Context for the operation (unused in memory store).
//   - template: *configv1.ServiceTemplate. The service template to save.
//
// Returns:
//   - error: Returns nil on success, or an error if the operation fails.
func (s *Store) SaveServiceTemplate(_ context.Context, template *configv1.ServiceTemplate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceTemplates[template.GetId()] = proto.Clone(template).(*configv1.ServiceTemplate)
	return nil
}

// DeleteServiceTemplate deletes a service template by ID.
//
// Summary: Removes a service template from memory.
//
// Parameters:
//   - ctx: context.Context. Context for the operation (unused in memory store).
//   - id: string. The unique identifier of the service template to delete.
//
// Returns:
//   - error: Returns nil on success, or an error if the operation fails.
func (s *Store) DeleteServiceTemplate(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.serviceTemplates, id)
	return nil
}
