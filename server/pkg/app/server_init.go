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
		if cfg != nil && (len(cfg.GetUpstreamServices()) > 0 || cfg.GlobalSettings != nil) {
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

	// Default Configuration
	defaultGS := &configv1.GlobalSettings{
		ProfileDefinitions: []*configv1.ProfileDefinition{
			{
				Name: proto.String("Default Dev"),
				Selector: &configv1.ProfileSelector{
					Tags: []string{"dev"},
				},
			},
		},
		DbPath: proto.String("mcpany.db"),
		Middlewares: []*configv1.Middleware{
			{Name: proto.String("auth"), Priority: proto.Int32(1), Disabled: proto.Bool(true)},
		},
	}

	// Default Weather Service for demonstration
	weatherService := &configv1.UpstreamServiceConfig{
		Id:   proto.String("weather-service"),
		Name: proto.String("weather-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
			CommandLineService: &configv1.CommandLineUpstreamService{
				Command: proto.String("echo"),
				Tools: []*configv1.ToolDefinition{
					{
						Name:        proto.String("get_weather"),
						Description: proto.String("Get current weather"),
						CallId:      proto.String("get_weather"),
					},
				},
				Calls: map[string]*configv1.CommandLineCallDefinition{
					"get_weather": {
						Args: []string{"{\"weather\": \"sunny\"}"},
					},
				},
				Resources: []*configv1.ResourceDefinition{
					{
						Uri:      proto.String("system://logs"),
						Name:     proto.String("System Logs"),
						MimeType: proto.String("text/plain"),
					},
				},
				Prompts: []*configv1.PromptDefinition{
					{
						Name:        proto.String("summarize_text"),
						Description: proto.String("Summarize text"),
					},
				},
			},
		},
	}

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
		b := make([]byte, 24) // 24 bytes = 32 chars in base64
		if _, err := rand.Read(b); err != nil {
			return fmt.Errorf("failed to generate random password: %w", err)
		}
		password = base64.URLEncoding.EncodeToString(b)
		logging.GetLogger().Info("****************************************************************")
		logging.GetLogger().Info("CRITICAL: Generated random admin password. Please change it immediately!", "password", password)
		logging.GetLogger().Info("****************************************************************")
	}

	hash, err := passhash.Password(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	adminUser := &configv1.User{
		Id: proto.String(username),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					Username:     proto.String(username),
					PasswordHash: proto.String(hash),
				},
			},
		},
		Roles: []string{"admin"},
	}

	if err := s.CreateUser(ctx, adminUser); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	logging.GetLogger().Info("Default admin user created successfully.", "username", username)
	return nil
}
