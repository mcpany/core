// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"google.golang.org/protobuf/proto"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	// Initialize Storage
	var store storage.Storage
	var closer func() error

	// Check for Postgres DSN
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = os.Getenv("POSTGRES_DSN")
	}

	if dsn != "" {
		return fmt.Errorf("postgres support disabled in seeder")
	}

	// Default to SQLite
	dbPath := "mcpany.db"
	if p := os.Getenv("DB_PATH"); p != "" {
		dbPath = p
	}
	db, err := sqlite.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open sqlite db: %w", err)
	}
	store = sqlite.NewStore(db)
	closer = db.Close

	defer func() {
		if closer != nil {
			_ = closer()
		}
	}()

	fmt.Println("Seeding database...")

	// 1. Create Admin User
	adminUser := configv1.User_builder{
		Id: proto.String("e2e-admin-core"),
		Roles: []string{"admin"},
		ProfileIds: []string{"dev", "prod"},
	}.Build()

	// For Opaque API, setting a OneOf usually requires using the specific setter on the message struct,
	// NOT the builder if the builder is experimental or limited.
	// OR we construct the message using the struct directly if possible, but Opaque API hides fields.
	// Since we are stuck with Opaque API + Builder issues, let's use the explicit setters provided by generated code.

	auth := configv1.Authentication_builder{}.Build()
	basicAuth := configv1.BasicAuth_builder{
		Username: proto.String("e2e-admin-core"),
		PasswordHash: proto.String("$2a$12$KPRtQETm7XKJP/L6FjYYxuCFpTK/oRs7v9U6hWx9XFnWy6UuDqK/a"),
	}.Build()

	// The generated code for OneOf `auth_method` usually provides `SetBasicAuth`.
	auth.SetBasicAuth(basicAuth)
	adminUser.SetAuthentication(auth)

	if err := store.CreateUser(ctx, adminUser); err != nil {
		fmt.Printf("Warning creating user: %v\n", err)
	} else {
		fmt.Println("Created user: e2e-admin-core")
	}

	// 2. Create Example Profile
	devProfile := configv1.ProfileDefinition_builder{
		Name:        proto.String("dev"),
	}.Build()
	if err := store.SaveProfile(ctx, devProfile); err != nil {
		fmt.Printf("Warning saving profile: %v\n", err)
	} else {
		fmt.Println("Created profile: dev")
	}

	// 3. Create Example Service (Echo)
	echoService := configv1.UpstreamServiceConfig_builder{
		Id:      proto.String("svc_echo"),
		Name:    proto.String("Echo Service"),
		Version: proto.String("v1.0"),
	}.Build()

	cmdSvc := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Tools: []*configv1.ToolDefinition{
			configv1.ToolDefinition_builder{
				Name:        proto.String("echo_tool"),
				Description: proto.String("Echoes back input"),
			}.Build(),
		},
	}.Build()

	echoService.SetCommandLineService(cmdSvc)

	if err := store.SaveService(ctx, echoService); err != nil {
		fmt.Printf("Warning saving service: %v\n", err)
	} else {
		fmt.Println("Created service: Echo Service")
	}

	// 4. Global Settings
	lvl := configv1.GlobalSettings_LOG_LEVEL_DEBUG
	settings := configv1.GlobalSettings_builder{
		LogLevel: &lvl,
	}.Build()

	if err := store.SaveGlobalSettings(ctx, settings); err != nil {
		fmt.Printf("Warning saving settings: %v\n", err)
	}

	fmt.Println("Seeding complete.")
	return nil
}
