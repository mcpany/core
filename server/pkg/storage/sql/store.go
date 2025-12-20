// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"fmt"
	"log/slog"

	"github.com/mcpany/core/pkg/config"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Store implements config.ServiceStore using Gorm.
type Store struct {
	db *gorm.DB
}

// NewStore creates a new SQL store.
// dialect can be "sqlite", "mysql", "postgres".
// dsn is the data source name (e.g. file path for sqlite).
func NewStore(dialect, dsn string) (*Store, error) {
	var dialector gorm.Dialector
	switch dialect {
	case "sqlite":
		dialector = sqlite.Open(dsn)
	// Add other dialects here when needed (mysql, postgres)
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}

	// Configure Gorm logger to use slog if needed, or silent.
	// For production, usually Silent or Error.
	// We can map slog.Level to Gorm log levels.
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	// AutoMigrate
	if err := db.AutoMigrate(&UpstreamService{}); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	return &Store{db: db}, nil
}

// Load retrieves and returns the McpAnyServerConfig.
func (s *Store) Load() (*configv1.McpAnyServerConfig, error) {
	services, err := s.ListServices()
	if err != nil {
		return nil, err
	}
	return &configv1.McpAnyServerConfig{
		UpstreamServices: services,
	}, nil
}

// SaveService saves an upstream service configuration.
func (s *Store) SaveService(service *configv1.UpstreamServiceConfig) error {
	if service.GetName() == "" {
		return fmt.Errorf("service name is required")
	}

	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(service)
	if err != nil {
		return fmt.Errorf("failed to marshal service config: %w", err)
	}

	id := service.GetId()
	if id == "" {
		id = service.GetName() // fallback
	}

	model := &UpstreamService{
		ID:     id,
		Name:   service.GetName(),
		Config: string(configJSON),
	}

	// Upsert based on ID (primary key)
	// We also want to ensure Name is unique.
	// Gorm's Save will update if ID exists.
	// However, if we change ID but keep Name, internal conflict?
	// Protobuf ID usually matches Name or UUID.
	// Let's use Save which performs upsert on PK.
	// We should mostly rely on Name as the logical ID for now?
	// The previous implementation used ID as PK and Name as Unique.

	// Use Clauses to handle upsert conflict on Name if ID is different?
	// Or just standard Save.
	result := s.db.Save(model)
	if result.Error != nil {
		return fmt.Errorf("failed to save service: %w", result.Error)
	}
	return nil
}

// GetService retrieves an upstream service configuration by name.
func (s *Store) GetService(name string) (*configv1.UpstreamServiceConfig, error) {
	var model UpstreamService
	result := s.db.Where("name = ?", name).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get service: %w", result.Error)
	}

	var service configv1.UpstreamServiceConfig
	if err := protojson.Unmarshal([]byte(model.Config), &service); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service config: %w", err)
	}
	// Restore ID/Name from model to ensure consistency?
	// Config JSON should have them.

	return &service, nil
}

// ListServices lists all upstream service configurations.
func (s *Store) ListServices() ([]*configv1.UpstreamServiceConfig, error) {
	var models []UpstreamService
	result := s.db.Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list services: %w", result.Error)
	}

	var services []*configv1.UpstreamServiceConfig
	for _, m := range models {
		var service configv1.UpstreamServiceConfig
		if err := protojson.Unmarshal([]byte(m.Config), &service); err != nil {
			slog.Error("Failed to unmarshal service config", "error", err, "id", m.ID)
			continue
		}
		services = append(services, &service)
	}
	return services, nil
}

// DeleteService deletes an upstream service configuration by name.
func (s *Store) DeleteService(name string) error {
	result := s.db.Where("name = ?", name).Delete(&UpstreamService{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete service: %w", result.Error)
	}
	return nil
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ensure Store implements config.ServiceStore
var _ config.ServiceStore = (*Store)(nil)
