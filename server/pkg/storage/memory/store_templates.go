// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// ListServiceTemplates _ is an unused parameter.
// Returns the result.
// Returns an error if the operation fails.
//
// Parameters:
//  - _ (context.Context): The _.
//
// Returns:
//  - []*configv1.ServiceTemplate: The result.
//  - error: Returns error on failure.
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
// _ is an unused parameter.
// id is the unique identifier.
// Returns the result.
// Returns an error if the operation fails.
//
// Parameters:
//  - _ (context.Context): The _.
//  - id (string): The unique identifier.
//
// Returns:
//  - *configv1.ServiceTemplate: The result.
//  - error: Returns error on failure.
func (s *Store) GetServiceTemplate(_ context.Context, id string) (*configv1.ServiceTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if t, ok := s.serviceTemplates[id]; ok {
		return proto.Clone(t).(*configv1.ServiceTemplate), nil
	}
	return nil, nil
}

// SaveServiceTemplate saves a service template.
// _ is an unused parameter.
// template is the template.
// Returns an error if the operation fails.
//
// Parameters:
//  - _ (context.Context): The _.
//  - template (*configv1.ServiceTemplate): The template.
//
// Returns:
//  - error: Returns error on failure.
func (s *Store) SaveServiceTemplate(_ context.Context, template *configv1.ServiceTemplate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceTemplates[template.GetId()] = proto.Clone(template).(*configv1.ServiceTemplate)
	return nil
}

// DeleteServiceTemplate deletes a service template by ID.
//
// Parameters:
//  - _ (context.Context): The _.
//  - id (string): The unique identifier.
//
// Returns:
//  - error: Returns error on failure.
func (s *Store) DeleteServiceTemplate(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.serviceTemplates, id)
	return nil
}
