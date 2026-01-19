// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
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

	log.Info("Database initialized successfully.")
	return nil
}
