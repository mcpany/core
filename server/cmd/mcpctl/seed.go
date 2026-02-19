// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/postgres"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func newSeedCmd() *cobra.Command {
	var dsn string

	cmd := &cobra.Command{
		Use:   "seed",
		Short: "Seed the database with initial data for testing",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			fmt.Printf("Connecting to database: %s\n", dsn)
			db, err := postgres.NewDB(dsn)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer db.Close()

			store := postgres.NewStore(db)

			// Truncate Tables
			fmt.Println("Truncating tables...")
			tables := []string{
				"upstream_services",
				"users",
				"global_settings",
				"secrets",
				"profile_definitions",
				"service_collections",
				"user_tokens",
			}
			for _, table := range tables {
				if _, err := db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
					// Ignore error if table doesn't exist
					fmt.Printf("Warning: failed to truncate %s: %v\n", table, err)
				}
			}

			// Seed Global Settings
			fmt.Println("Seeding Global Settings...")
			logLevel := configv1.GlobalSettings_LOG_LEVEL_INFO
			settings := configv1.GlobalSettings_builder{
				LogLevel: &logLevel,
			}.Build()
			if err := store.SaveGlobalSettings(ctx, settings); err != nil {
				return fmt.Errorf("failed to seed global settings: %w", err)
			}

			// Seed User
			fmt.Println("Seeding User...")
			user := configv1.User_builder{
				Id:    proto.String("test-user"),
				Roles: []string{"admin"},
			}.Build()
			if err := store.CreateUser(ctx, user); err != nil {
				return fmt.Errorf("failed to seed user: %w", err)
			}

			// Seed Service
			fmt.Println("Seeding Service...")

			inputSchema, err := structpb.NewStruct(map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{"type": "string"},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to create input schema: %w", err)
			}

			service := configv1.UpstreamServiceConfig_builder{
				Id:      proto.String("test-service"),
				Name:    proto.String("test-service"),
				Version: proto.String("1.0.0"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: proto.String("http://localhost:8081"),
					Tools: []*configv1.ToolDefinition{
						configv1.ToolDefinition_builder{
							Name:        proto.String("echo"),
							Description: proto.String("Echoes back the input"),
							InputSchema: inputSchema,
						}.Build(),
					},
				}.Build(),
			}.Build()

			if err := store.SaveService(ctx, service); err != nil {
				return fmt.Errorf("failed to seed service: %w", err)
			}

			// Seed Profile
			fmt.Println("Seeding Profile...")
			profile := configv1.ProfileDefinition_builder{
				Name: proto.String("default"),
			}.Build()
			if err := store.SaveProfile(ctx, profile); err != nil {
				return fmt.Errorf("failed to seed profile: %w", err)
			}

			// Seed Secret
			fmt.Println("Seeding Secret...")
			secret := configv1.Secret_builder{
				Id:    proto.String("test-secret"),
				Name:  proto.String("Test Secret"),
				Key:   proto.String("TEST_SECRET"),
				Value: proto.String("supersecret"),
			}.Build()
			if err := store.SaveSecret(ctx, secret); err != nil {
				return fmt.Errorf("failed to seed secret: %w", err)
			}

			fmt.Println("Database seeded successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&dsn, "dsn", "postgres://mcpany:secret@localhost:5432/mcpany?sslmode=disable", "Database connection string")

	return cmd
}
