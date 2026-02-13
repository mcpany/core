// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// BuiltinTemplates contains the seed configurations for high-value MCP servers.
// Deprecated: Use BuiltinServiceTemplates instead.
var BuiltinTemplates []*configv1.UpstreamServiceConfig

// BuiltinServiceTemplates contains the rich seed templates for the wizard.
var BuiltinServiceTemplates []*configv1.ServiceTemplate

func init() {
	// Legacy templates (CLI-based)
	BuiltinTemplates = []*configv1.UpstreamServiceConfig{
		mkTemplate(
			"github-cli",
			"GitHub (CLI)",
			`{
  "type": "object",
  "title": "GitHub Configuration",
  "properties": {
    "GITHUB_PERSONAL_ACCESS_TOKEN": {
      "type": "string",
      "title": "Personal Access Token",
      "description": "A GitHub PAT with repo permissions.",
      "format": "password"
    }
  },
  "required": ["GITHUB_PERSONAL_ACCESS_TOKEN"]
}`,
			"npx -y @modelcontextprotocol/server-github",
		),
	}

	// Rich Service Templates (The new standard)
	BuiltinServiceTemplates = []*configv1.ServiceTemplate{
		// 1. Google Calendar (OpenAPI)
		configv1.ServiceTemplate_builder{
			Id:          proto.String("google-calendar"),
			Name:        proto.String("Google Calendar"),
			Description: proto.String("Manage events and calendars via Google API."),
			Icon:        proto.String("calendar"),
			Tags:        []string{"google", "productivity", "calendar"},
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("google-calendar"),
				OpenapiService: configv1.OpenapiUpstreamService_builder{
					SpecUrl: proto.String("https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/googleapis.com/calendar/v3/openapi.yaml"),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					Oauth2: configv1.OAuth2Auth_builder{
						TokenUrl:         proto.String("https://oauth2.googleapis.com/token"),
						AuthorizationUrl: proto.String("https://accounts.google.com/o/oauth2/v2/auth"),
						Scopes:           proto.String("https://www.googleapis.com/auth/calendar"),
					}.Build(),
				}.Build(),
				ConfigurationSchema: proto.String(`{
					"type": "object",
					"title": "Google Calendar Config",
					"properties": {},
					"description": "Uses OAuth2. No additional config required."
				}`),
			}.Build(),
		}.Build(),

		// 2. GitHub (OpenAPI)
		configv1.ServiceTemplate_builder{
			Id:          proto.String("github"),
			Name:        proto.String("GitHub"),
			Description: proto.String("Interact with repositories, issues, and PRs."),
			Icon:        proto.String("github"),
			Tags:        []string{"development", "git", "scm"},
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("github"),
				OpenapiService: configv1.OpenapiUpstreamService_builder{
					SpecUrl: proto.String("https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.yaml"),
					Address: proto.String("https://api.github.com"),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					BearerToken: configv1.BearerTokenAuth_builder{
						Token: configv1.SecretValue_builder{
							PlainText: proto.String("${GITHUB_TOKEN}"),
						}.Build(),
					}.Build(),
				}.Build(),
				ConfigurationSchema: proto.String(`{
					"type": "object",
					"title": "GitHub Configuration",
					"properties": {
						"GITHUB_TOKEN": {
							"type": "string",
							"title": "Personal Access Token",
							"format": "password"
						}
					},
					"required": ["GITHUB_TOKEN"]
				}`),
			}.Build(),
		}.Build(),

		// 3. Filesystem (Internal Native)
		configv1.ServiceTemplate_builder{
			Id:          proto.String("filesystem"),
			Name:        proto.String("Local Filesystem"),
			Description: proto.String("Secure access to local directories."),
			Icon:        proto.String("folder"),
			Tags:        []string{"system", "storage"},
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("filesystem"),
				FilesystemService: configv1.FilesystemUpstreamService_builder{
					RootPaths: map[string]string{"/data": "./data"},
					ReadOnly:  proto.Bool(false),
					Os:        configv1.OsFs_builder{}.Build(),
				}.Build(),
				ConfigurationSchema: proto.String(`{
					"type": "object",
					"title": "Filesystem Config",
					"properties": {
						"ROOT_PATH": { "type": "string", "default": "./data" }
					}
				}`),
			}.Build(),
		}.Build(),

		// 4. Postgres (Internal Native)
		configv1.ServiceTemplate_builder{
			Id:          proto.String("postgres"),
			Name:        proto.String("PostgreSQL"),
			Description: proto.String("Database access and query execution."),
			Icon:        proto.String("database"),
			Tags:        []string{"database", "sql"},
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("postgres"),
				SqlService: configv1.SqlUpstreamService_builder{
					Driver: proto.String("postgres"),
					Dsn:    proto.String("${POSTGRES_DSN}"),
				}.Build(),
				ConfigurationSchema: proto.String(`{
					"type": "object",
					"title": "PostgreSQL Config",
					"properties": {
						"POSTGRES_DSN": {
							"type": "string",
							"title": "Connection String",
							"default": "postgres://user:pass@localhost:5432/db"
						}
					},
					"required": ["POSTGRES_DSN"]
				}`),
			}.Build(),
		}.Build(),

		// 5. Linear (OpenAPI)
		configv1.ServiceTemplate_builder{
			Id:          proto.String("linear"),
			Name:        proto.String("Linear"),
			Description: proto.String("Issue tracking and project management."),
			Icon:        proto.String("linear"),
			Tags:        []string{"pm", "issues"},
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("linear"),
				OpenapiService: configv1.OpenapiUpstreamService_builder{
					SpecUrl: proto.String("https://raw.githubusercontent.com/linear/linear/master/api/openapi.yaml"),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					ApiKey: configv1.APIKeyAuth_builder{
						Value: configv1.SecretValue_builder{
							PlainText: proto.String("${LINEAR_API_KEY}"),
						}.Build(),
					}.Build(),
				}.Build(),
				ConfigurationSchema: proto.String(`{
					"type": "object",
					"title": "Linear Config",
					"properties": {
						"LINEAR_API_KEY": { "type": "string", "format": "password" }
					},
					"required": ["LINEAR_API_KEY"]
				}`),
			}.Build(),
		}.Build(),
	}
}

func mkTemplate(id, name, schema, command string) *configv1.UpstreamServiceConfig {
	t := configv1.UpstreamServiceConfig_builder{
		Id:                  proto.String(id),
		Name:                proto.String(name),
		Version:             proto.String("1.0.0"),
		ConfigurationSchema: proto.String(schema),
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Command: proto.String(command),
			Env:     make(map[string]*configv1.SecretValue),
		}.Build(),
		AutoDiscoverTool: proto.Bool(true),
	}.Build()
	return t
}
