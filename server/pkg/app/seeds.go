// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"google.golang.org/protobuf/proto"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// BuiltinTemplates contains the seed configurations for high-value MCP servers.
var BuiltinTemplates []*configv1.UpstreamServiceConfig

// BuiltinServiceTemplates contains the rich service templates.
var BuiltinServiceTemplates []*configv1.ServiceTemplate

func init() {
	BuiltinTemplates = []*configv1.UpstreamServiceConfig{
		mkTemplate(
			"github",
			"GitHub",
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
		mkTemplate(
			"postgres",
			"PostgreSQL",
			`{
  "type": "object",
  "title": "PostgreSQL Configuration",
  "properties": {
    "POSTGRES_URL": {
      "type": "string",
      "title": "Connection URL",
      "description": "postgresql://user:password@host:port/dbname",
      "default": "postgresql://postgres:postgres@localhost:5432/postgres"
    }
  },
  "required": ["POSTGRES_URL"]
}`,
			"npx -y @modelcontextprotocol/server-postgres ${POSTGRES_URL}",
		),
		mkTemplate(
			"filesystem",
			"Filesystem",
			`{
  "type": "object",
  "title": "Filesystem Configuration",
  "properties": {
    "ALLOWED_PATHS": {
      "type": "string",
      "title": "Allowed Paths",
      "description": "Space-separated list of directories the server can access.",
      "default": "."
    }
  },
  "required": ["ALLOWED_PATHS"]
}`,
			"npx -y @modelcontextprotocol/server-filesystem ${ALLOWED_PATHS}",
		),
		mkTemplate(
			"brave-search",
			"Brave Search",
			`{
  "type": "object",
  "title": "Brave Search Configuration",
  "properties": {
    "BRAVE_API_KEY": {
      "type": "string",
      "title": "API Key",
      "description": "Your Brave Search API Key.",
      "format": "password"
    }
  },
  "required": ["BRAVE_API_KEY"]
}`,
			"npx -y @modelcontextprotocol/server-brave-search",
		),
		mkTemplate(
			"google-maps",
			"Google Maps",
			`{
  "type": "object",
  "title": "Google Maps Configuration",
  "properties": {
    "GOOGLE_MAPS_API_KEY": {
      "type": "string",
      "title": "API Key",
      "description": "Your Google Maps API Key.",
      "format": "password"
    }
  },
  "required": ["GOOGLE_MAPS_API_KEY"]
}`,
			"npx -y @modelcontextprotocol/server-google-maps",
		),
		mkTemplate(
			"slack",
			"Slack",
			`{
  "type": "object",
  "title": "Slack Configuration",
  "properties": {
    "SLACK_BOT_TOKEN": {
      "type": "string",
      "title": "Bot Token",
      "description": "xoxb-...",
      "format": "password"
    },
    "SLACK_TEAM_ID": {
      "type": "string",
      "title": "Team ID",
      "description": "T..."
    }
  },
  "required": ["SLACK_BOT_TOKEN", "SLACK_TEAM_ID"]
}`,
			"npx -y @modelcontextprotocol/server-slack",
		),
		mkTemplate(
			"sentry",
			"Sentry",
			`{
  "type": "object",
  "title": "Sentry Configuration",
  "properties": {
    "SENTRY_AUTH_TOKEN": {
      "type": "string",
      "title": "Auth Token",
      "description": "Sentry Authentication Token.",
      "format": "password"
    }
  },
  "required": ["SENTRY_AUTH_TOKEN"]
}`,
			"npx -y @modelcontextprotocol/server-sentry",
		),
		mkTemplate(
			"memory",
			"Memory",
			`{
  "type": "object",
  "title": "Memory Configuration",
  "properties": {},
  "description": "Knowledge graph memory server. No configuration required."
}`,
			"npx -y @modelcontextprotocol/server-memory",
		),
		mkTemplate(
			"gitlab",
			"GitLab",
			`{
  "type": "object",
  "title": "GitLab Configuration",
  "properties": {
    "GITLAB_PERSONAL_ACCESS_TOKEN": {
      "type": "string",
      "title": "Personal Access Token",
      "description": "GitLab PAT.",
      "format": "password"
    },
    "GITLAB_API_URL": {
      "type": "string",
      "title": "API URL",
      "description": "Base URL for GitLab API (optional, defaults to gitlab.com).",
      "default": "https://gitlab.com/api/v4"
    }
  },
  "required": ["GITLAB_PERSONAL_ACCESS_TOKEN"]
}`,
			"npx -y @modelcontextprotocol/server-gitlab",
		),
		mkTemplate(
			"cloudflare",
			"Cloudflare",
			`{
  "type": "object",
  "title": "Cloudflare Configuration",
  "properties": {
    "CLOUDFLARE_API_TOKEN": {
      "type": "string",
      "title": "API Token",
      "description": "Cloudflare API Token.",
      "format": "password"
    },
    "CLOUDFLARE_ACCOUNT_ID": {
      "type": "string",
      "title": "Account ID",
      "description": "Cloudflare Account ID."
    }
  },
  "required": ["CLOUDFLARE_API_TOKEN", "CLOUDFLARE_ACCOUNT_ID"]
}`,
			"npx -y @modelcontextprotocol/server-cloudflare",
		),
	}

	BuiltinServiceTemplates = []*configv1.ServiceTemplate{
		configv1.ServiceTemplate_builder{
			Id:          proto.String("google-calendar"),
			Name:        proto.String("Google Calendar"),
			Description: proto.String("Manage events and calendars."),
			Icon:        proto.String("calendar"),
			Tags:        []string{"google", "productivity"},
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("google_calendar"),
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
				OpenapiService: configv1.OpenapiUpstreamService_builder{
					SpecUrl: proto.String("https://api.apis.guru/v2/specs/googleapis.com/calendar/v3/openapi.yaml"),
				}.Build(),
			}.Build(),
		}.Build(),
		configv1.ServiceTemplate_builder{
			Id:          proto.String("github"),
			Name:        proto.String("GitHub"),
			Description: proto.String("Interact with repositories, issues, and PRs."),
			Icon:        proto.String("github"),
			Tags:        []string{"dev", "git"},
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("github"),
				McpService: configv1.McpUpstreamService_builder{
					HttpConnection: configv1.McpStreamableHttpConnection_builder{
						HttpAddress: proto.String("https://api.github.com"),
					}.Build(),
					ToolAutoDiscovery: proto.Bool(true),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					BearerToken: configv1.BearerTokenAuth_builder{
						Token: configv1.SecretValue_builder{
							PlainText: proto.String(""),
						}.Build(),
					}.Build(),
				}.Build(),
				OpenapiService: configv1.OpenapiUpstreamService_builder{
					Address: proto.String("https://api.github.com"),
					SpecUrl: proto.String("https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.yaml"),
				}.Build(),
			}.Build(),
		}.Build(),
		configv1.ServiceTemplate_builder{
			Id:          proto.String("linear"),
			Name:        proto.String("Linear"),
			Description: proto.String("Issue tracking and project management."),
			Icon:        proto.String("linear"),
			Tags:        []string{"dev", "pm"},
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("linear"),
				McpService: configv1.McpUpstreamService_builder{
					HttpConnection: configv1.McpStreamableHttpConnection_builder{
						HttpAddress: proto.String("https://api.linear.app/graphql"),
					}.Build(),
					ToolAutoDiscovery: proto.Bool(true),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					ApiKey: configv1.APIKeyAuth_builder{
						Value: configv1.SecretValue_builder{PlainText: proto.String("")}.Build(),
					}.Build(),
				}.Build(),
				OpenapiService: configv1.OpenapiUpstreamService_builder{
					SpecUrl: proto.String("https://raw.githubusercontent.com/linear/linear/master/api/openapi.yaml"),
				}.Build(),
			}.Build(),
		}.Build(),
	}
}

func mkTemplate(id, name, schema, command string) *configv1.UpstreamServiceConfig {
	t := &configv1.UpstreamServiceConfig{}
	t.SetId(id)
	t.SetName(name)
	t.SetVersion("1.0.0")
	t.SetConfigurationSchema(schema)

	cmd := &configv1.CommandLineUpstreamService{}
	cmd.SetCommand(command)
	cmd.SetEnv(make(map[string]*configv1.SecretValue))

	t.SetCommandLineService(cmd)
	t.SetAutoDiscoverTool(true)
	return t
}
