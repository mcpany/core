// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// BuiltinTemplates contains the seed configurations for high-value MCP servers.
var BuiltinTemplates []*configv1.UpstreamServiceConfig

// BuiltinServiceTemplates contains the seed configurations for the Service Wizard (HTTP/OAuth/Rich Metadata).
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
		mkServiceTemplate(
			"empty",
			"Custom Service",
			"Configure a service from scratch.",
			"server",
			[]string{"other"},
			func() *configv1.UpstreamServiceConfig {
				c := &configv1.UpstreamServiceConfig{}
				// Initialize with empty HTTP service to default the form type
				h := &configv1.HttpUpstreamService{}
				h.SetAddress("")
				c.SetHttpService(h)
				return c
			}(),
		),
		mkServiceTemplate(
			"google-calendar",
			"Google Calendar",
			"Calendar management.",
			"calendar",
			[]string{"productivity", "google"},
			func() *configv1.UpstreamServiceConfig {
				c := &configv1.UpstreamServiceConfig{}
				c.SetName("google-calendar")

				conn := &configv1.McpStreamableHttpConnection{}
				conn.SetHttpAddress("https://calendar.googleapis.com")

				mcp := &configv1.McpUpstreamService{}
				mcp.SetHttpConnection(conn)
				mcp.SetToolAutoDiscovery(true)
				c.SetMcpService(mcp)

				auth := &configv1.Authentication{}
				oauth := &configv1.OAuth2Auth{}
				oauth.SetScopes("https://www.googleapis.com/auth/calendar")
				auth.SetOauth2(oauth)
				c.SetUpstreamAuth(auth)

				c.SetConfigurationSchema(`{
					"type": "object",
					"title": "Google Calendar Configuration",
					"description": "Connect to Google Calendar via OAuth.",
					"properties": {},
					"required": []
				}`)
				return c
			}(),
		),
		mkServiceTemplate(
			"github",
			"GitHub",
			"Code hosting and collaboration.",
			"github",
			[]string{"development", "git"},
			func() *configv1.UpstreamServiceConfig {
				c := &configv1.UpstreamServiceConfig{}
				c.SetName("github")

				conn := &configv1.McpStreamableHttpConnection{}
				conn.SetHttpAddress("https://api.github.com")

				mcp := &configv1.McpUpstreamService{}
				mcp.SetHttpConnection(conn)
				mcp.SetToolAutoDiscovery(true)
				c.SetMcpService(mcp)

				auth := &configv1.Authentication{}
				oauth := &configv1.OAuth2Auth{}
				oauth.SetScopes("repo,user")
				auth.SetOauth2(oauth)
				c.SetUpstreamAuth(auth)

				c.SetConfigurationSchema(`{
					"type": "object",
					"title": "GitHub Configuration",
					"description": "Connect to GitHub via OAuth.",
					"properties": {},
					"required": []
				}`)
				return c
			}(),
		),
		mkServiceTemplate(
			"gitlab",
			"GitLab",
			"DevOps lifecycle tool.",
			"gitlab",
			[]string{"development", "git"},
			func() *configv1.UpstreamServiceConfig {
				c := &configv1.UpstreamServiceConfig{}
				c.SetName("gitlab")

				conn := &configv1.McpStreamableHttpConnection{}
				conn.SetHttpAddress("https://gitlab.com/api/v4")

				mcp := &configv1.McpUpstreamService{}
				mcp.SetHttpConnection(conn)
				mcp.SetToolAutoDiscovery(true)
				c.SetMcpService(mcp)

				auth := &configv1.Authentication{}
				oauth := &configv1.OAuth2Auth{}
				oauth.SetScopes("api")
				auth.SetOauth2(oauth)
				c.SetUpstreamAuth(auth)

				c.SetConfigurationSchema(`{
					"type": "object",
					"title": "GitLab Configuration",
					"description": "Connect to GitLab via OAuth.",
					"properties": {},
					"required": []
				}`)
				return c
			}(),
		),
		mkServiceTemplate(
			"slack",
			"Slack",
			"Team communication and collaboration.",
			"slack",
			[]string{"productivity", "chat"},
			func() *configv1.UpstreamServiceConfig {
				c := &configv1.UpstreamServiceConfig{}
				c.SetName("slack")

				conn := &configv1.McpStreamableHttpConnection{}
				conn.SetHttpAddress("https://slack.com/api")

				mcp := &configv1.McpUpstreamService{}
				mcp.SetHttpConnection(conn)
				mcp.SetToolAutoDiscovery(true)
				c.SetMcpService(mcp)

				auth := &configv1.Authentication{}
				oauth := &configv1.OAuth2Auth{}
				oauth.SetScopes("channels:read,chat:write,files:read")
				auth.SetOauth2(oauth)
				c.SetUpstreamAuth(auth)

				c.SetConfigurationSchema(`{
					"type": "object",
					"title": "Slack Configuration",
					"description": "Connect to Slack via OAuth.",
					"properties": {},
					"required": []
				}`)
				return c
			}(),
		),
		mkServiceTemplate(
			"notion",
			"Notion",
			"All-in-one workspace for notes and docs.",
			"notion",
			[]string{"productivity", "docs"},
			func() *configv1.UpstreamServiceConfig {
				c := &configv1.UpstreamServiceConfig{}
				c.SetName("notion")

				conn := &configv1.McpStreamableHttpConnection{}
				conn.SetHttpAddress("https://api.notion.com/v1")

				mcp := &configv1.McpUpstreamService{}
				mcp.SetHttpConnection(conn)
				mcp.SetToolAutoDiscovery(true)
				c.SetMcpService(mcp)

				auth := &configv1.Authentication{}
				oauth := &configv1.OAuth2Auth{}
				oauth.SetScopes("basic")
				auth.SetOauth2(oauth)
				c.SetUpstreamAuth(auth)

				c.SetConfigurationSchema(`{
					"type": "object",
					"title": "Notion Configuration",
					"description": "Connect to Notion via OAuth.",
					"properties": {},
					"required": []
				}`)
				return c
			}(),
		),
		mkServiceTemplate(
			"linear",
			"Linear",
			"Issue tracking and project management.",
			"linear",
			[]string{"development", "pm"},
			func() *configv1.UpstreamServiceConfig {
				c := &configv1.UpstreamServiceConfig{}
				c.SetName("linear")

				conn := &configv1.McpStreamableHttpConnection{}
				conn.SetHttpAddress("https://api.linear.app/graphql")

				mcp := &configv1.McpUpstreamService{}
				mcp.SetHttpConnection(conn)
				mcp.SetToolAutoDiscovery(true)
				c.SetMcpService(mcp)

				auth := &configv1.Authentication{}
				oauth := &configv1.OAuth2Auth{}
				oauth.SetScopes("read,write")
				auth.SetOauth2(oauth)
				c.SetUpstreamAuth(auth)

				c.SetConfigurationSchema(`{
					"type": "object",
					"title": "Linear Configuration",
					"description": "Connect to Linear via OAuth.",
					"properties": {},
					"required": []
				}`)
				return c
			}(),
		),
		mkServiceTemplate(
			"jira",
			"Jira",
			"Issue tracking and agile project management.",
			"jira",
			[]string{"development", "pm"},
			func() *configv1.UpstreamServiceConfig {
				c := &configv1.UpstreamServiceConfig{}
				c.SetName("jira")

				conn := &configv1.McpStreamableHttpConnection{}
				conn.SetHttpAddress("https://api.atlassian.com/ex/jira")

				mcp := &configv1.McpUpstreamService{}
				mcp.SetHttpConnection(conn)
				mcp.SetToolAutoDiscovery(true)
				c.SetMcpService(mcp)

				auth := &configv1.Authentication{}
				oauth := &configv1.OAuth2Auth{}
				oauth.SetScopes("read:jira-work,write:jira-work")
				auth.SetOauth2(oauth)
				c.SetUpstreamAuth(auth)

				c.SetConfigurationSchema(`{
					"type": "object",
					"title": "Jira Configuration",
					"description": "Connect to Jira via OAuth.",
					"properties": {},
					"required": []
				}`)
				return c
			}(),
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

func mkServiceTemplate(id, name, desc, icon string, tags []string, config *configv1.UpstreamServiceConfig) *configv1.ServiceTemplate {
	t := &configv1.ServiceTemplate{}
	t.SetId(id)
	t.SetName(name)
	t.SetDescription(desc)
	t.SetIcon(icon)
	t.SetTags(tags)
	t.SetServiceConfig(config)
	return t
}
