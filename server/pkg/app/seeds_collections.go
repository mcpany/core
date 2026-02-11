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

// BuiltinCollections contains the seed configurations for official service collections.
var BuiltinCollections []*configv1.Collection

func init() {
	BuiltinCollections = []*configv1.Collection{
		configv1.Collection_builder{
			Name:        proto.String("Data Engineering Stack"),
			Description: proto.String("Essential tools for data pipelines (PostgreSQL, Filesystem, Python)"),
			Version:     proto.String("1.0.0"),
			Services: []*configv1.UpstreamServiceConfig{
				configv1.UpstreamServiceConfig_builder{
					Id:            proto.String("sqlite-db"),
					Name:          proto.String("SQLite Database"),
					Version:       proto.String("1.0.0"),
					SanitizedName: proto.String("sqlite-db"),
					CommandLineService: configv1.CommandLineUpstreamService_builder{
						Command: proto.String("npx -y @modelcontextprotocol/server-sqlite"),
						Env: map[string]*configv1.SecretValue{
							"DB_PATH": configv1.SecretValue_builder{
								PlainText: proto.String("./data.db"),
							}.Build(),
						},
					}.Build(),
					AutoDiscoverTool: proto.Bool(false),
				}.Build(),
			},
		}.Build(),
		configv1.Collection_builder{
			Name:        proto.String("Web Dev Assistant"),
			Description: proto.String("GitHub, Browser, and Terminal tools for web development."),
			Version:     proto.String("1.0.0"),
			Services: []*configv1.UpstreamServiceConfig{
				configv1.UpstreamServiceConfig_builder{
					Id:            proto.String("github"),
					Name:          proto.String("GitHub"),
					Version:       proto.String("1.0.0"),
					SanitizedName: proto.String("github"),
					CommandLineService: configv1.CommandLineUpstreamService_builder{
						Command: proto.String("npx -y @modelcontextprotocol/server-github"),
						Env: map[string]*configv1.SecretValue{
							"GITHUB_PERSONAL_ACCESS_TOKEN": configv1.SecretValue_builder{
								PlainText: proto.String(""),
							}.Build(),
						},
					}.Build(),
					AutoDiscoverTool: proto.Bool(true),
				}.Build(),
			},
		}.Build(),
	}
}

func (a *Application) seedCollections(ctx context.Context, store config.Store) error {
	log := logging.GetLogger()
	s, ok := store.(storage.Storage)
	if !ok {
		return nil
	}

	for _, col := range BuiltinCollections {
		existing, err := s.GetServiceCollection(ctx, col.GetName())
		if err != nil {
			log.Error("Failed to check for existing collection", "name", col.GetName(), "error", err)
			continue
		}
		if existing == nil {
			log.Info("Seeding collection", "name", col.GetName())
			if err := s.SaveServiceCollection(ctx, col); err != nil {
				return fmt.Errorf("failed to save seeded collection %s: %w", col.GetName(), err)
			}
		}
	}
	return nil
}
