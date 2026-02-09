// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package servicetemplates provides seeding functionality for service templates.
package servicetemplates

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage"
)

// Seeder seeds the database with service templates.
type Seeder struct {
	Store       storage.Storage
	ExamplesDir string
}

// ConfigFile represents the structure of the config.yaml in examples.
type ConfigFile struct {
	UpstreamServices []map[string]any `yaml:"upstream_services"`
}

// Seed walks the examples directory and saves service templates.
func (s *Seeder) Seed(ctx context.Context) error {
	// ⚡ BOLT: Cleanup - Removed dead code that scanned directories but didn't persist anything.
	// The implementation now strictly uses the built-in templates as intended.

	// Hardcoded list to match client.ts quality
	templates := s.getBuiltInTemplates()
	for _, t := range templates {
		if err := s.Store.SaveServiceTemplate(ctx, t); err != nil {
			return fmt.Errorf("failed to save template %s: %w", t.GetId(), err)
		}
	}

	return nil
}

func (s *Seeder) getBuiltInTemplates() []*configv1.ServiceTemplate {
	return []*configv1.ServiceTemplate{
		configv1.ServiceTemplate_builder{
			Id:          proto.String("google-calendar"),
			Name:        proto.String("Google Calendar"),
			Description: proto.String("Calendar management."),
			Icon:        proto.String("google-calendar"),
			Tags:        []string{"productivity", "google"},
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("google-calendar"),
				McpService: configv1.McpUpstreamService_builder{
					HttpConnection: configv1.McpStreamableHttpConnection_builder{
						HttpAddress: proto.String("https://calendar.googleapis.com"),
					}.Build(),
					ToolAutoDiscovery: proto.Bool(true),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					Oauth2: configv1.OAuth2Auth_builder{
						Scopes: proto.String("https://www.googleapis.com/auth/calendar"),
					}.Build(),
				}.Build(),
			}.Build(),
		}.Build(),
		configv1.ServiceTemplate_builder{
			Id:          proto.String("github"),
			Name:        proto.String("GitHub"),
			Description: proto.String("Code hosting and collaboration."),
			Icon:        proto.String("github"),
			Tags:        []string{"development", "git"},
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("github"),
				McpService: configv1.McpUpstreamService_builder{
					HttpConnection: configv1.McpStreamableHttpConnection_builder{
						HttpAddress: proto.String("https://api.github.com"),
					}.Build(),
					ToolAutoDiscovery: proto.Bool(true),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					Oauth2: configv1.OAuth2Auth_builder{
						Scopes: proto.String("repo,user"),
					}.Build(),
				}.Build(),
			}.Build(),
		}.Build(),
		configv1.ServiceTemplate_builder{
			Id:          proto.String("gitlab"),
			Name:        proto.String("GitLab"),
			Description: proto.String("DevOps lifecycle tool."),
			Icon:        proto.String("gitlab"),
			Tags:        []string{"development", "git"},
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("gitlab"),
				McpService: configv1.McpUpstreamService_builder{
					HttpConnection: configv1.McpStreamableHttpConnection_builder{
						HttpAddress: proto.String("https://gitlab.com/api/v4"),
					}.Build(),
					ToolAutoDiscovery: proto.Bool(true),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					Oauth2: configv1.OAuth2Auth_builder{
						Scopes: proto.String("api"),
					}.Build(),
				}.Build(),
			}.Build(),
		}.Build(),
		configv1.ServiceTemplate_builder{
			Id:          proto.String("slack"),
			Name:        proto.String("Slack"),
			Description: proto.String("Team communication and collaboration."),
			Icon:        proto.String("slack"),
			Tags:        []string{"productivity", "chat"},
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("slack"),
				McpService: configv1.McpUpstreamService_builder{
					HttpConnection: configv1.McpStreamableHttpConnection_builder{
						HttpAddress: proto.String("https://slack.com/api"),
					}.Build(),
					ToolAutoDiscovery: proto.Bool(true),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					Oauth2: configv1.OAuth2Auth_builder{
						Scopes: proto.String("channels:read,chat:write,files:read"),
					}.Build(),
				}.Build(),
			}.Build(),
		}.Build(),
		configv1.ServiceTemplate_builder{
			Id:          proto.String("notion"),
			Name:        proto.String("Notion"),
			Description: proto.String("All-in-one workspace for notes and docs."),
			Icon:        proto.String("notion"),
			Tags:        []string{"productivity", "docs"},
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("notion"),
				McpService: configv1.McpUpstreamService_builder{
					HttpConnection: configv1.McpStreamableHttpConnection_builder{
						HttpAddress: proto.String("https://api.notion.com/v1"),
					}.Build(),
					ToolAutoDiscovery: proto.Bool(true),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					Oauth2: configv1.OAuth2Auth_builder{
						Scopes: proto.String("basic"),
					}.Build(),
				}.Build(),
			}.Build(),
		}.Build(),
		configv1.ServiceTemplate_builder{
			Id:          proto.String("linear"),
			Name:        proto.String("Linear"),
			Description: proto.String("Issue tracking and project management."),
			Icon:        proto.String("linear"),
			Tags:        []string{"development", "pm"},
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("linear"),
				McpService: configv1.McpUpstreamService_builder{
					HttpConnection: configv1.McpStreamableHttpConnection_builder{
						HttpAddress: proto.String("https://api.linear.app/graphql"),
					}.Build(),
					ToolAutoDiscovery: proto.Bool(true),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					Oauth2: configv1.OAuth2Auth_builder{
						Scopes: proto.String("read,write"),
					}.Build(),
				}.Build(),
			}.Build(),
		}.Build(),
		configv1.ServiceTemplate_builder{
			Id:          proto.String("jira"),
			Name:        proto.String("Jira"),
			Description: proto.String("Issue tracking and agile project management."),
			Icon:        proto.String("jira"),
			Tags:        []string{"development", "pm"},
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("jira"),
				McpService: configv1.McpUpstreamService_builder{
					HttpConnection: configv1.McpStreamableHttpConnection_builder{
						HttpAddress: proto.String("https://api.atlassian.com/ex/jira"),
					}.Build(),
					ToolAutoDiscovery: proto.Bool(true),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					Oauth2: configv1.OAuth2Auth_builder{
						Scopes: proto.String("read:jira-work,write:jira-work"),
					}.Build(),
				}.Build(),
			}.Build(),
		}.Build(),
	}
}
