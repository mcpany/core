// Copyright 2026 Author(s) of MCP Any
// ListServiceTemplates retrieves all service templates.
//
// Summary: Lists all stored service templates.
//
// Parameters:
//   - _: context.Context. Unused.
//
// Returns:
// GetServiceTemplate retrieves a service template by ID.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Retrieves a service template by ID.
//
// Parameters:
//   - _: context.Context. Unused.
// SaveServiceTemplate saves a service template.
//
// Summary: Stores a service template.
//
// Parameters:
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - _: context.Context. Unused.
// DeleteServiceTemplate deletes a service template by ID.
//
// Summary: Deletes a service template.
//
// Parameters:
//   - _: context.Context. Unused.
//   - id: string. The template ID.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Removes from the internal service template map.
// Errors:
//   - triggers relevant error states on failure.
package memory

package memory

func (s *Store) DeleteServiceTemplate(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.serviceTemplates, id)
	return nil
}
