// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// CreateSystemWebhook creates a new system webhook.
func (s *Store) CreateSystemWebhook(ctx context.Context, webhook *configv1.SystemWebhook) error {
	if webhook.GetId() == "" {
		return fmt.Errorf("webhook ID is required")
	}

	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(webhook)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook: %w", err)
	}

	query := `
	INSERT INTO system_webhooks (id, config_json, updated_at)
	VALUES (?, ?, CURRENT_TIMESTAMP)
	`
	_, err = s.db.ExecContext(ctx, query, webhook.GetId(), string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to create system webhook: %w", err)
	}
	return nil
}

// ListSystemWebhooks retrieves all system webhooks.
func (s *Store) ListSystemWebhooks(ctx context.Context) ([]*configv1.SystemWebhook, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM system_webhooks")
	if err != nil {
		return nil, fmt.Errorf("failed to query system webhooks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var webhooks []*configv1.SystemWebhook
	for rows.Next() {
		var configJSON []byte
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan webhook config: %w", err)
		}

		var w configv1.SystemWebhook
		if err := protojson.Unmarshal(configJSON, &w); err != nil {
			return nil, fmt.Errorf("failed to unmarshal webhook: %w", err)
		}
		webhooks = append(webhooks, &w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return webhooks, nil
}

// GetSystemWebhook retrieves a system webhook by ID.
func (s *Store) GetSystemWebhook(ctx context.Context, id string) (*configv1.SystemWebhook, error) {
	query := "SELECT config_json FROM system_webhooks WHERE id = ?"
	row := s.db.QueryRowContext(ctx, query, id)

	var configJSON []byte
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan webhook config: %w", err)
	}

	var w configv1.SystemWebhook
	if err := protojson.Unmarshal(configJSON, &w); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook: %w", err)
	}
	return &w, nil
}

// UpdateSystemWebhook updates an existing system webhook.
func (s *Store) UpdateSystemWebhook(ctx context.Context, webhook *configv1.SystemWebhook) error {
	if webhook.GetId() == "" {
		return fmt.Errorf("webhook ID is required")
	}

	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(webhook)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook: %w", err)
	}

	query := `
	UPDATE system_webhooks
	SET config_json = ?, updated_at = CURRENT_TIMESTAMP
	WHERE id = ?
	`
	res, err := s.db.ExecContext(ctx, query, string(configJSON), webhook.GetId())
	if err != nil {
		return fmt.Errorf("failed to update system webhook: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("system webhook not found")
	}
	return nil
}

// DeleteSystemWebhook deletes a system webhook by ID.
func (s *Store) DeleteSystemWebhook(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM system_webhooks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete system webhook: %w", err)
	}
	return nil
}
