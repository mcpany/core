// Package main provides a utility to seed the MCP Any database with initial data for testing.
// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"google.golang.org/protobuf/proto"
)

func main() {
	dbPath := flag.String("db-path", "data/mcpany.db", "Path to the SQLite database file")
	skillsDir := flag.String("skills-dir", "data/skills", "Path to the skills directory")
	flag.Parse()

	ctx := context.Background()

	// 1. Seed Database
	if err := seedDatabase(ctx, *dbPath); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// 2. Seed Skills
	if err := seedSkills(*skillsDir); err != nil {
		log.Fatalf("Failed to seed skills: %v", err)
	}

	fmt.Println("Successfully seeded data for documentation and tests.")
}

func seedDatabase(ctx context.Context, dbPath string) error {
	db, err := sqlite.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Printf("Warning: failed to close database: %v\n", err)
		}
	}()

	store := sqlite.NewStore(db)

	// Services
	services := []*configv1.UpstreamServiceConfig{
		{
			Id:   proto.String("postgres-primary"),
			Name: proto.String("Primary DB"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String("http://postgres:5432"),
				},
			},
			Version: proto.String("1.0.0"),
		},
		{
			Id:   proto.String("openai-gateway"),
			Name: proto.String("OpenAI Gateway"),
			ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
				McpService: &configv1.McpUpstreamService{
					ConnectionType: &configv1.McpUpstreamService_HttpConnection{
						HttpConnection: &configv1.McpStreamableHttpConnection{
							HttpAddress: proto.String("http://openai-mcp:8080"),
						},
					},
				},
			},
			Version: proto.String("2.1.0"),
		},
		{
			Id:   proto.String("payment-gateway"),
			Name: proto.String("Payment Gateway"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String("https://stripe.com"),
				},
			},
			Version: proto.String("v1.2.0"),
		},
		{
			Id:   proto.String("user-service"),
			Name: proto.String("User Service"),
			ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
				GrpcService: &configv1.GrpcUpstreamService{
					Address: proto.String("localhost:50051"),
				},
			},
			Version: proto.String("v1.0"),
		},
	}
	for _, s := range services {
		if err := store.SaveService(ctx, s); err != nil {
			return fmt.Errorf("failed to save service %s: %w", s.GetName(), err)
		}
	}

	// Settings
	level := configv1.GlobalSettings_LOG_LEVEL_INFO
	settings := &configv1.GlobalSettings{
		LogLevel: &level,
	}
	if err := store.SaveGlobalSettings(ctx, settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	// Users
	user := &configv1.User{
		Id:    proto.String("admin-user"),
		Roles: []string{"admin"},
	}
	if err := store.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Secrets
	secret := &configv1.Secret{
		Id:   proto.String("sec-1"),
		Name: proto.String("API_KEY"),
		Key:  proto.String("mcp-api-key"),
	}
	if err := store.SaveSecret(ctx, secret); err != nil {
		return fmt.Errorf("failed to save secret: %w", err)
	}

	// Profiles
	profiles := []*configv1.ProfileDefinition{
		{
			Name: proto.String("production"),
		},
		{
			Name: proto.String("staging"),
		},
	}
	for _, p := range profiles {
		if err := store.SaveProfile(ctx, p); err != nil {
			return fmt.Errorf("failed to save profile %s: %w", p.GetName(), err)
		}
	}

	// Collections
	collections := []*configv1.Collection{
		{
			Name: proto.String("ai-stack"),
			Services: []*configv1.UpstreamServiceConfig{
				services[0], services[1],
			},
		},
	}
	for _, c := range collections {
		if err := store.SaveServiceCollection(ctx, c); err != nil {
			return fmt.Errorf("failed to save collection %s: %w", c.GetName(), err)
		}
	}

	// Credentials
	creds := []*configv1.Credential{
		{
			Id:   proto.String("cred-1"),
			Name: proto.String("Default Postgres Creds"),
			Authentication: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_BasicAuth{
					BasicAuth: &configv1.BasicAuth{
						Username: proto.String("admin"),
					},
				},
			},
		},
	}
	for _, c := range creds {
		if err := store.SaveCredential(ctx, c); err != nil {
			return fmt.Errorf("failed to save credential %s: %w", c.GetName(), err)
		}
	}

	return nil
}

func seedSkills(skillsDir string) error {
	if err := os.MkdirAll(skillsDir, 0750); err != nil {
		return err
	}

	// Sample Skill: "system-monitor"
	skillName := "system-monitor"
	skillPath := filepath.Join(skillsDir, skillName)
	if err := os.MkdirAll(skillPath, 0750); err != nil {
		return err
	}

	skillContent := `---
name: system-monitor
description: Monitors system health and alerts on failures.
version: 1.0.0
---

# System Monitor Skill

This skill provides tools for monitoring system resource usage and health.
`
	if err := os.WriteFile(filepath.Join(skillPath, "SKILL.md"), []byte(skillContent), 0600); err != nil {
		return err
	}

	return nil
}
