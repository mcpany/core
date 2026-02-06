// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// BuiltinTemplates contains the seed configurations for high-value MCP servers.
var BuiltinTemplates []*configv1.UpstreamServiceConfig

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
		mkTemplate(
			"linear",
			"Linear",
			`{
  "type": "object",
  "title": "Linear Configuration",
  "properties": {
    "LINEAR_API_KEY": {
      "type": "string",
      "title": "API Key",
      "description": "Your Linear API Key.",
      "format": "password"
    }
  },
  "required": ["LINEAR_API_KEY"]
}`,
			"npx -y @modelcontextprotocol/server-linear",
		),
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
