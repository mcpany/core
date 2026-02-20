// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// Service Templates

// ListServiceTemplates retrieves all service templates from the in-memory store.
//
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//
// Returns:
//  []*configv1.ServiceTemplate: A slice of service templates.
//  error: Always nil.
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  id (string): The unique identifier of the service template.
//
// Returns:
//  *configv1.ServiceTemplate: The service template if found.
//  error: Always nil (returns nil result if not found).
func (s *Store) GetServiceTemplate(_ context.Context, id string) (*configv1.ServiceTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if t, ok := s.serviceTemplates[id]; ok {
		return proto.Clone(t).(*configv1.ServiceTemplate), nil
	}
	return nil, nil
}

// SaveServiceTemplate saves or updates a service template.
//
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  template (*configv1.ServiceTemplate): The service template to save.
//
// Returns:
//  error: Always nil.
//
// Side Effects:
//  Updates the internal map of service templates.
func (s *Store) SaveServiceTemplate(_ context.Context, template *configv1.ServiceTemplate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceTemplates[template.GetId()] = proto.Clone(template).(*configv1.ServiceTemplate)
	return nil
}

// DeleteServiceTemplate deletes a service template by ID.
//
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  id (string): The unique identifier of the service template to delete.
//
// Returns:
//  error: Always nil.
//
// Side Effects:
//  Removes the service template from the internal map.
func (s *Store) DeleteServiceTemplate(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.serviceTemplates, id)
	return nil
}
