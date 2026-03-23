// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/util/passhash"
)

func (a *Application) initializeDatabase(ctx context.Context, store config.Store, cfg *configv1.McpAnyServerConfig) error {
	log := logging.GetLogger()

	// If the runtime config already contains explicit content, don't seed DB defaults over it.
	if cfg != nil {
		hasServices := len(cfg.GetUpstreamServices()) > 0
		hasUsers := len(cfg.GetUsers()) > 0
		hasCollections := len(cfg.GetCollections()) > 0
		hasGlobalSettings := hasConfiguredFields(cfg.GetGlobalSettings())
		if hasServices || hasUsers || hasCollections || hasGlobalSettings {
			log.Debug("Configuration already present (detected in merged config), skipping database initialization.")
			return nil
		}
	}

	// Double-check if already initialized in DB specifically
	s, ok := store.(storage.Storage)
	if !ok {
		// Just Load using Store interface
		dbCfg, err := store.Load(ctx)
		if err != nil {
			return err
		}
		if dbCfg != nil && (len(dbCfg.GetUpstreamServices()) > 0 || dbCfg.GetGlobalSettings() != nil) {
			return nil // Already initialized in DB
		}
	} else {
		// Use Storage interface
		services, err := s.ListServices(ctx)
		if err != nil {
			return err
		}
		if len(services) > 0 {
			return nil
		}
		// Also check global settings?
		gs, err := s.GetGlobalSettings(ctx)
		if err == nil && gs != nil {
			return nil
		}
	}

	log.Info("Database appears empty and no configuration provided, initializing with default configuration...")

	// Default Configuration
	defaultGS := configv1.GlobalSettings_builder{
		ProfileDefinitions: []*configv1.ProfileDefinition{
			configv1.ProfileDefinition_builder{
				Name: proto.String("default"),
				Selector: configv1.ProfileSelector_builder{
					Tags: []string{"dev"},
				}.Build(),
			}.Build(),
		},
		DbPath: proto.String("mcpany.db"),
		Middlewares: []*configv1.Middleware{
			configv1.Middleware_builder{
				Name:     proto.String("auth"),
				Priority: proto.Int32(1),
				Disabled: proto.Bool(false),
			}.Build(),
		},
	}.Build()
	// panic("DEBUG: initializeDatabase called") // Commented out to avoid crashing, but using error log as panic alternative if needed.
	// actually, use fmt.Println to bypass logger if logger is borked
	fmt.Println("DEBUG:fmt: Initializing DB with defaultGS")
	log.Info("DEBUG: Initializing DB with defaultGS", "middlewares", defaultGS.GetMiddlewares())

	// Default Weather Service for demonstration
	weatherService := configv1.UpstreamServiceConfig_builder{
		Id:   proto.String("weather-service"),
		Name: proto.String("weather-service"),
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Command: proto.String("echo"),
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:        proto.String("get_weather"),
					Description: proto.String("Get current weather"),
					CallId:      proto.String("get_weather"),
				}.Build(),
			},
			Calls: map[string]*configv1.CommandLineCallDefinition{
				"get_weather": configv1.CommandLineCallDefinition_builder{
					Args: []string{"{\"weather\": \"sunny\"}"},
				}.Build(),
			},
			Resources: []*configv1.ResourceDefinition{
				configv1.ResourceDefinition_builder{
					Uri:      proto.String("system://logs"),
					Name:     proto.String("System Logs"),
					MimeType: proto.String("text/plain"),
				}.Build(),
			},
			Prompts: []*configv1.PromptDefinition{
				configv1.PromptDefinition_builder{
					Name:        proto.String("summarize_text"),
					Description: proto.String("Summarize text"),
				}.Build(),
			},
		}.Build(),
	}.Build()

	// Save to DB
	if s, ok := store.(storage.Storage); ok {
		if err := s.SaveGlobalSettings(ctx, defaultGS); err != nil {
			return fmt.Errorf("failed to save default global settings: %w", err)
		}
		if err := s.SaveService(ctx, weatherService); err != nil {
			return fmt.Errorf("failed to save default weather service: %w", err)
		}
	} else {
		log.Warn("Store/Storage does not support saving defaults.")
	}

	// Initialize Service Templates
	if err := a.seedTemplates(ctx, store); err != nil {
		log.Error("Failed to seed service templates", "error", err)
	}

	// Initialize Service Collections
	if err := a.seedCollections(ctx, store); err != nil {
		log.Error("Failed to seed service collections", "error", err)
	}

	// Initialize Admin User
	if err := a.initializeAdminUser(ctx, store); err != nil {
		log.Error("Failed to initialize admin user", "error", err)
		// We don't fail hard here to allow server to start, but auth might be broken for admin
	}

	log.Info("Database initialized successfully.")
	return nil
}

func hasConfiguredFields(message proto.Message) bool {
	if message == nil {
		return false
	}

	hasFields := false
	message.ProtoReflect().Range(func(protoreflect.FieldDescriptor, protoreflect.Value) bool {
		hasFields = true
		return false
	})

	return hasFields
}

func (a *Application) seedCollections(ctx context.Context, store config.Store) error {
	s, ok := store.(storage.Storage)
	if !ok {
		return nil
	}

	// Check if collections already exist
	collections, err := s.ListServiceCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	if len(collections) > 0 {
		return nil
	}

	logging.GetLogger().Info("Seeding builtin service collections...", "count", len(BuiltinServiceCollections))

	for _, c := range BuiltinServiceCollections {
		if err := s.SaveServiceCollection(ctx, c); err != nil {
			return fmt.Errorf("failed to save collection %s: %w", c.GetName(), err)
		}
	}

	return nil
}

func (a *Application) initializeAdminUser(ctx context.Context, store config.Store) error {
	s, ok := store.(storage.Storage)
	if !ok {
		return nil // Cannot list/save users
	}

	users, err := s.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	if len(users) > 0 {
		return nil // Users already exist
	}

	logging.GetLogger().Info("No users found, creating default admin user...")

	username := os.Getenv("MCPANY_ADMIN_INIT_USERNAME")
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("MCPANY_ADMIN_INIT_PASSWORD")
	if password == "" {
		// Generate strong random password
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return fmt.Errorf("failed to generate random password: %w", err)
		}
		password = base64.RawURLEncoding.EncodeToString(b)
		logging.GetLogger().Warn("⚠️  GENERATED ADMIN PASSWORD: " + password + " ⚠️")
		logging.GetLogger().Warn("Please save this password immediately and change it upon first login.")
	}

	hash, err := passhash.Password(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	adminUser := configv1.User_builder{
		Id: proto.String(username),
		Authentication: configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				Username:     proto.String(username),
				PasswordHash: proto.String(hash),
			}.Build(),
		}.Build(),
		Roles: []string{"admin"},
	}.Build()

	if err := s.CreateUser(ctx, adminUser); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	logging.GetLogger().Info("Default admin user created successfully.", "username", username)
	return nil
}

func (a *Application) seedTemplates(ctx context.Context, store config.Store) error {
	s, ok := store.(storage.Storage)
	if !ok {
		return nil
	}

	// Check if templates already exist
	templates, err := s.ListServiceTemplates(ctx)
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	if len(templates) > 0 {
		return nil
	}

	logging.GetLogger().Info("Seeding builtin service templates...", "count", len(BuiltinServiceTemplates))

	for _, t := range BuiltinServiceTemplates {
		if err := s.SaveServiceTemplate(ctx, t); err != nil {
			return fmt.Errorf("failed to save template %s: %w", t.GetId(), err)
		}
	}

	// Seed Swarm Topology Mock Data for Dashboard Widget
	swarmTopologyData := `{
		"nodes": [
			{ "id": "n1", "label": "Primary Orchestrator", "type": "validator", "status": "locked", "x": 50, "y": 50 },
			{ "id": "n2", "label": "Research Agent", "type": "agent", "status": "active", "x": 20, "y": 30 },
			{ "id": "n3", "label": "Tool Exec", "type": "service", "status": "idle", "x": 20, "y": 70 },
			{ "id": "n4", "label": "Synthesizer", "type": "agent", "status": "active", "x": 80, "y": 50 },
			{ "id": "n5", "label": "Rogue Node", "type": "agent", "status": "stall", "x": 80, "y": 20 }
		],
		"edges": [
			{ "source": "n2", "target": "n1", "status": "healthy", "hash": "0x1A4" },
			{ "source": "n1", "target": "n3", "status": "healthy", "hash": "0x2B9" },
			{ "source": "n1", "target": "n4", "status": "healthy", "hash": "0x3C1" },
			{ "source": "n5", "target": "n1", "status": "blocked", "hash": "INVALID_GRAFT" }
		],
		"anomalies": ["ARI Hub: Logic Graft Blocked from Rogue Node (n5)"]
	}`
	if err := s.SaveMockData(ctx, "swarm-topology", swarmTopologyData); err != nil {
		logging.GetLogger().Error("failed to seed swarm topology mock data", "error", err)
	}

	return nil
}
