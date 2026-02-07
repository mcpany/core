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

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/util/passhash"
)

func (a *Application) initializeDatabase(ctx context.Context, store config.Store) error {
	log := logging.GetLogger()
	// Check if already initialized
	s, ok := store.(storage.Storage)
	if !ok {
		// Just Load using Store interface
		cfg, err := store.Load(ctx)
		if err != nil {
			return err
		}
		if cfg != nil && (len(cfg.GetUpstreamServices()) > 0 || cfg.GetGlobalSettings() != nil) {
			return nil // Already initialized
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

	log.Info("Database appears empty, initializing with default configuration...")

	// Default Dashboard Layout (matches frontend defaults)
	defaultLayout := `[
		{"instanceId":"metrics-1","type":"metrics","title":"Metrics Overview","size":"full","hidden":false},
		{"instanceId":"quick-actions-1","type":"quick-actions","title":"Quick Actions","size":"third","hidden":false},
		{"instanceId":"service-health-1","type":"service-health","title":"Service Health","size":"third","hidden":false},
		{"instanceId":"failure-rate-1","type":"failure-rate","title":"Tool Failure Rates","size":"third","hidden":false},
		{"instanceId":"activity-1","type":"recent-activity","title":"Recent Activity","size":"half","hidden":false},
		{"instanceId":"uptime-1","type":"uptime","title":"System Uptime","size":"half","hidden":false},
		{"instanceId":"request-volume-1","type":"request-volume","title":"Request Volume","size":"half","hidden":false},
		{"instanceId":"top-tools-1","type":"top-tools","title":"Top Tools","size":"third","hidden":false}
	]`

	// Default Configuration
	defaultGS := configv1.GlobalSettings_builder{
		ProfileDefinitions: []*configv1.ProfileDefinition{
			configv1.ProfileDefinition_builder{
				Name: proto.String("Default Dev"),
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
		Dashboard: configv1.DashboardConfig_builder{
			LayoutJson: proto.String(defaultLayout),
		}.Build(),
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

	// Initialize Admin User
	if err := a.initializeAdminUser(ctx, store); err != nil {
		log.Error("Failed to initialize admin user", "error", err)
		// We don't fail hard here to allow server to start, but auth might be broken for admin
	}

	log.Info("Database initialized successfully.")
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
