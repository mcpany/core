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
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - []*configv1.ServiceTemplate: A list of service templates.
//   - error: An error if the operation fails.
func (s *Store) ListServiceTemplates(_ context.Context) ([]*configv1.ServiceTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.ServiceTemplate, 0, len(s.serviceTemplates))
	for _, t := range s.serviceTemplates {
		list = append(list, proto.Clone(t).(*configv1.ServiceTemplate))
	}
	return list, nil
}

// GetServiceTemplate retrieves a service template by name.
//
// Summary: Fetches a service template by name.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - name: string. The name of the template.
//
// Returns:
//   - *configv1.ServiceTemplate: The service template, or nil if not found.
//   - error: An error if the operation fails.
func (s *Store) GetServiceTemplate(_ context.Context, id string) (*configv1.ServiceTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if t, ok := s.serviceTemplates[id]; ok {
		return proto.Clone(t).(*configv1.ServiceTemplate), nil
	}
	return nil, nil
}

// SaveServiceTemplate persists a service template.
//
// Summary: Saves or updates a service template.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - template: *configv1.ServiceTemplate. The template to save.
//
// Returns:
//   - error: An error if the operation fails.
func (s *Store) SaveServiceTemplate(_ context.Context, template *configv1.ServiceTemplate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceTemplates[template.GetId()] = proto.Clone(template).(*configv1.ServiceTemplate)
	return nil
}

// DeleteServiceTemplate removes a service template.
//
// Summary: Deletes a service template by name.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - name: string. The name of the template.
//
// Returns:
//   - error: An error if the operation fails.
func (s *Store) DeleteServiceTemplate(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.serviceTemplates, id)
	return nil
}
