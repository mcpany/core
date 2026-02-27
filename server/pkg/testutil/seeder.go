// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"database/sql"
	"fmt"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Seed populates the database with deterministic test data.
func Seed(ctx context.Context, db *sql.DB) error {
	// 1. Seed Global Settings
	settings := configv1.GlobalSettings_builder{
		ApiKey:   proto.String("test-api-key"),
		LogLevel: configv1.GlobalSettings_LOG_LEVEL_INFO.Enum(),
	}.Build()

	if err := saveGlobalSettings(ctx, db, settings); err != nil {
		return fmt.Errorf("failed to seed global settings: %w", err)
	}

	// 2. Seed User
	// Note: The User proto definition does not include a 'name' field, only 'id' and 'roles'.
	userJSON := `{
		"id": "test-user",
		"roles": ["admin"]
	}`

	user := &configv1.User{}
	if err := protojson.Unmarshal([]byte(userJSON), user); err != nil {
		return fmt.Errorf("failed to unmarshal user json: %w", err)
	}

	if err := saveUser(ctx, db, user); err != nil {
		return fmt.Errorf("failed to seed user: %w", err)
	}

	// 3. Seed Service
	serviceJSON := `{
		"id": "test-service",
		"name": "test-service",
		"version": "1.0.0",
		"http_service": {
			"address": "http://localhost:8080"
		}
	}`
	service := &configv1.UpstreamServiceConfig{}
	if err := protojson.Unmarshal([]byte(serviceJSON), service); err != nil {
		return fmt.Errorf("failed to unmarshal service json: %w", err)
	}

	if err := saveService(ctx, db, service); err != nil {
		return fmt.Errorf("failed to seed service: %w", err)
	}

	return nil
}

func saveGlobalSettings(ctx context.Context, db *sql.DB, settings *configv1.GlobalSettings) error {
	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(settings)
	if err != nil {
		return err
	}
	query := `INSERT OR REPLACE INTO global_settings (id, config_json, updated_at) VALUES (1, ?, CURRENT_TIMESTAMP)`
	_, err = db.ExecContext(ctx, query, string(configJSON))
	return err
}

func saveUser(ctx context.Context, db *sql.DB, user *configv1.User) error {
	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(user)
	if err != nil {
		return err
	}
	query := `INSERT OR REPLACE INTO users (id, config_json, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)`
	_, err = db.ExecContext(ctx, query, user.GetId(), string(configJSON))
	return err
}

func saveService(ctx context.Context, db *sql.DB, service *configv1.UpstreamServiceConfig) error {
	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(service)
	if err != nil {
		return err
	}
	// ID and Name are used; schema has ID primary key, Name unique.
	// We'll insert based on ID.
	query := `INSERT OR REPLACE INTO upstream_services (id, name, config_json, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)`
	id := service.GetId()
	if id == "" {
		id = service.GetName()
	}
	_, err = db.ExecContext(ctx, query, id, service.GetName(), string(configJSON))
	return err
}
