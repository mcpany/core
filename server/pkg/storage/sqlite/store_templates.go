// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// Service Templates

// ListServiceTemplates serves as a public interface for interacting with ListServiceTemplates.
//
// Summary: List the service templates appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Store) ListServiceTemplates(ctx context.Context) ([]*configv1.ServiceTemplate, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM service_templates")
	if err != nil {
		return nil, fmt.Errorf("failed to query service_templates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var templates []*configv1.ServiceTemplate
	for rows.Next() {
		var configJSON []byte
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan config_json: %w", err)
		}

		var template configv1.ServiceTemplate
		if err := protojson.Unmarshal(configJSON, &template); err != nil {
			return nil, fmt.Errorf("failed to unmarshal service template: %w", err)
		}
		templates = append(templates, &template)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	return templates, nil
}

// GetServiceTemplate serves as a public interface for interacting with GetServiceTemplate.
//
// Summary: Fetches and returns the underlying service template from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Store) GetServiceTemplate(ctx context.Context, id string) (*configv1.ServiceTemplate, error) {
	query := "SELECT config_json FROM service_templates WHERE id = $1"
	row := s.db.QueryRowContext(ctx, query, id)

	var configJSON []byte
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan config_json: %w", err)
	}

	var template configv1.ServiceTemplate
	if err := protojson.Unmarshal(configJSON, &template); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service template: %w", err)
	}
	return &template, nil
}

// SaveServiceTemplate serves as a public interface for interacting with SaveServiceTemplate.
//
// Summary: Save the service template appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Store) SaveServiceTemplate(ctx context.Context, template *configv1.ServiceTemplate) error {
	if template.GetId() == "" {
		return fmt.Errorf("template ID is required")
	}

	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal service template: %w", err)
	}

	query := `
	INSERT INTO service_templates (id, name, config_json, updated_at)
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		config_json = excluded.config_json,
		updated_at = excluded.updated_at;
	`
	_, err = s.db.ExecContext(ctx, query, template.GetId(), template.GetName(), string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to save service template: %w", err)
	}
	return nil
}

// DeleteServiceTemplate serves as a public interface for interacting with DeleteServiceTemplate.
//
// Summary: Delete the service template appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Store) DeleteServiceTemplate(ctx context.Context, id string) error {
	query := "DELETE FROM service_templates WHERE id = $1"
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete service template: %w", err)
	}
	return nil
}
