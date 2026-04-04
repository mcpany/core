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
// Summary: Retrieves all service templates.
//
// Parameters:
//   - unnamed (context.Context): Parameter.
//
// Returns:
//   - []*configv1.ServiceTemplate: Return value.
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
//
// Side Effects:
//   - None.

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
//   - unnamed (context.Context): Parameter.
//   - id (string): Parameter.
//
// Returns:
//   - *configv1.ServiceTemplate: Return value.
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
//
// Side Effects:
//   - None.

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
// Summary: Saves a service template.
//
// Parameters:
//   - unnamed (context.Context): Parameter.
//   - template (*configv1.ServiceTemplate): Parameter.
//
// Returns:
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
//
// Side Effects:
//   - None.

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
//   - unnamed (context.Context): Parameter.
//   - id (string): Parameter.
//
// Returns:
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
//
// Side Effects:
//   - None.

func (s *Store) DeleteServiceTemplate(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.serviceTemplates, id)
	return nil
}
