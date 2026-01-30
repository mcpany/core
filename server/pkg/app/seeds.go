// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// BuiltinTemplates contains the seed data for the service templates.
var BuiltinTemplates []*configv1.UpstreamServiceConfig

func init() {
	// GitHub
	github := &configv1.UpstreamServiceConfig{}
	github.SetId("github")
	github.SetName("GitHub")
	github.SetVersion("1.0.0")
	github.SetTags([]string{"git", "scm", "microsoft"})
	github.SetConfigurationSchema(`{
  "type": "object",
  "title": "GitHub Configuration",
  "required": ["GITHUB_PERSONAL_ACCESS_TOKEN"],
  "properties": {
    "GITHUB_PERSONAL_ACCESS_TOKEN": {
      "type": "string",
      "title": "Personal Access Token",
      "description": "A GitHub PAT with repo permissions.",
      "format": "password"
    }
  }
}`)

	githubCmd := &configv1.CommandLineUpstreamService{}
	githubCmd.SetCommand("npx -y @modelcontextprotocol/server-github")
	githubCmd.SetEnv(map[string]*configv1.SecretValue{
		"GITHUB_PERSONAL_ACCESS_TOKEN": {
			Value: &configv1.SecretValue_PlainText{PlainText: ""},
		},
	})
	github.SetCommandLineService(githubCmd)

	githubMcp := &configv1.McpUpstreamService{}
	githubMcp.SetToolAutoDiscovery(true)
	github.SetMcpService(githubMcp)

	BuiltinTemplates = append(BuiltinTemplates, github)

	// PostgreSQL
	postgres := &configv1.UpstreamServiceConfig{}
	postgres.SetId("postgres")
	postgres.SetName("PostgreSQL")
	postgres.SetVersion("1.0.0")
	postgres.SetTags([]string{"sql", "db", "relational"})
	postgres.SetConfigurationSchema(`{
  "type": "object",
  "title": "PostgreSQL Configuration",
  "required": ["POSTGRES_URL"],
  "properties": {
    "POSTGRES_URL": {
      "type": "string",
      "title": "Connection URL",
      "description": "postgresql://user:password@localhost:5432/dbname",
      "default": "postgresql://postgres:postgres@localhost:5432/postgres"
    }
  }
}`)

	postgresCmd := &configv1.CommandLineUpstreamService{}
	postgresCmd.SetCommand("sh -c 'npx -y @modelcontextprotocol/server-postgres \"$POSTGRES_URL\"'")
	postgresCmd.SetEnv(map[string]*configv1.SecretValue{
		"POSTGRES_URL": {
			Value: &configv1.SecretValue_PlainText{PlainText: ""},
		},
	})
	postgres.SetCommandLineService(postgresCmd)

	postgresMcp := &configv1.McpUpstreamService{}
	postgresMcp.SetToolAutoDiscovery(true)
	postgres.SetMcpService(postgresMcp)

	BuiltinTemplates = append(BuiltinTemplates, postgres)

	// Filesystem
	fs := &configv1.UpstreamServiceConfig{}
	fs.SetId("filesystem")
	fs.SetName("Filesystem")
	fs.SetVersion("1.0.0")
	fs.SetTags([]string{"fs", "local", "io"})
	fs.SetConfigurationSchema(`{
  "type": "object",
  "title": "Filesystem Configuration",
  "required": ["ALLOWED_PATHS"],
  "properties": {
    "ALLOWED_PATHS": {
      "type": "string",
      "title": "Allowed Paths",
      "description": "Space-separated list of directories to expose.",
      "default": "./"
    }
  }
}`)

	fsCmd := &configv1.CommandLineUpstreamService{}
	fsCmd.SetCommand("sh -c 'npx -y @modelcontextprotocol/server-filesystem $ALLOWED_PATHS'")
	fsCmd.SetEnv(map[string]*configv1.SecretValue{
		"ALLOWED_PATHS": {
			Value: &configv1.SecretValue_PlainText{PlainText: ""},
		},
	})
	fs.SetCommandLineService(fsCmd)

	fsMcp := &configv1.McpUpstreamService{}
	fsMcp.SetToolAutoDiscovery(true)
	fs.SetMcpService(fsMcp)

	BuiltinTemplates = append(BuiltinTemplates, fs)

	// Brave Search
	brave := &configv1.UpstreamServiceConfig{}
	brave.SetId("brave-search")
	brave.SetName("Brave Search")
	brave.SetVersion("1.0.0")
	brave.SetTags([]string{"search", "api", "web"})
	brave.SetConfigurationSchema(`{
  "type": "object",
  "title": "Brave Search Configuration",
  "required": ["BRAVE_API_KEY"],
  "properties": {
    "BRAVE_API_KEY": {
      "type": "string",
      "title": "API Key",
      "description": "Your Brave Search API Key.",
      "format": "password"
    }
  }
}`)

	braveCmd := &configv1.CommandLineUpstreamService{}
	braveCmd.SetCommand("npx -y @modelcontextprotocol/server-brave-search")
	braveCmd.SetEnv(map[string]*configv1.SecretValue{
		"BRAVE_API_KEY": {
			Value: &configv1.SecretValue_PlainText{PlainText: ""},
		},
	})
	brave.SetCommandLineService(braveCmd)

	braveMcp := &configv1.McpUpstreamService{}
	braveMcp.SetToolAutoDiscovery(true)
	brave.SetMcpService(braveMcp)

	BuiltinTemplates = append(BuiltinTemplates, brave)

	// Slack
	slack := &configv1.UpstreamServiceConfig{}
	slack.SetId("slack")
	slack.SetName("Slack")
	slack.SetVersion("1.0.0")
	slack.SetTags([]string{"chat", "messaging", "work"})
	slack.SetConfigurationSchema(`{
  "type": "object",
  "title": "Slack Configuration",
  "required": ["SLACK_BOT_TOKEN", "SLACK_TEAM_ID"],
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
      "description": "T01234567"
    }
  }
}`)

	slackCmd := &configv1.CommandLineUpstreamService{}
	slackCmd.SetCommand("npx -y @modelcontextprotocol/server-slack")
	slackCmd.SetEnv(map[string]*configv1.SecretValue{
		"SLACK_BOT_TOKEN": {
			Value: &configv1.SecretValue_PlainText{PlainText: ""},
		},
		"SLACK_TEAM_ID": {
			Value: &configv1.SecretValue_PlainText{PlainText: ""},
		},
	})
	slack.SetCommandLineService(slackCmd)

	slackMcp := &configv1.McpUpstreamService{}
	slackMcp.SetToolAutoDiscovery(true)
	slack.SetMcpService(slackMcp)

	BuiltinTemplates = append(BuiltinTemplates, slack)

	// Memory
	mem := &configv1.UpstreamServiceConfig{}
	mem.SetId("memory")
	mem.SetName("Memory")
	mem.SetVersion("1.0.0")
	mem.SetTags([]string{"graph", "knowledge", "experimental"})
	mem.SetConfigurationSchema(`{
  "type": "object",
  "title": "Memory Configuration",
  "properties": {}
}`)

	memCmd := &configv1.CommandLineUpstreamService{}
	memCmd.SetCommand("npx -y @modelcontextprotocol/server-memory")
	mem.SetCommandLineService(memCmd)

	memMcp := &configv1.McpUpstreamService{}
	memMcp.SetToolAutoDiscovery(true)
	mem.SetMcpService(memMcp)

	BuiltinTemplates = append(BuiltinTemplates, mem)

	// Sentry
	sentry := &configv1.UpstreamServiceConfig{}
	sentry.SetId("sentry")
	sentry.SetName("Sentry")
	sentry.SetVersion("1.0.0")
	sentry.SetTags([]string{"monitoring", "error-tracking"})
	sentry.SetConfigurationSchema(`{
  "type": "object",
  "title": "Sentry Configuration",
  "required": ["SENTRY_AUTH_TOKEN"],
  "properties": {
    "SENTRY_AUTH_TOKEN": {
      "type": "string",
      "title": "Auth Token",
      "description": "Sentry Auth Token with project:read access.",
      "format": "password"
    }
  }
}`)

	sentryCmd := &configv1.CommandLineUpstreamService{}
	sentryCmd.SetCommand("npx -y @modelcontextprotocol/server-sentry")
	sentryCmd.SetEnv(map[string]*configv1.SecretValue{
		"SENTRY_AUTH_TOKEN": {
			Value: &configv1.SecretValue_PlainText{PlainText: ""},
		},
	})
	sentry.SetCommandLineService(sentryCmd)

	sentryMcp := &configv1.McpUpstreamService{}
	sentryMcp.SetToolAutoDiscovery(true)
	sentry.SetMcpService(sentryMcp)

	BuiltinTemplates = append(BuiltinTemplates, sentry)

	// SQLite
	sqlite := &configv1.UpstreamServiceConfig{}
	sqlite.SetId("sqlite")
	sqlite.SetName("SQLite")
	sqlite.SetVersion("1.0.0")
	sqlite.SetTags([]string{"sql", "db", "local"})
	sqlite.SetConfigurationSchema(`{
  "type": "object",
  "title": "SQLite Configuration",
  "required": ["SQLITE_FILE"],
  "properties": {
    "SQLITE_FILE": {
      "type": "string",
      "title": "Database File",
      "description": "Path to the SQLite database file.",
      "default": "test.db"
    }
  }
}`)

	sqliteCmd := &configv1.CommandLineUpstreamService{}
	sqliteCmd.SetCommand("sh -c 'npx -y @modelcontextprotocol/server-sqlite \"$SQLITE_FILE\"'")
	sqliteCmd.SetEnv(map[string]*configv1.SecretValue{
		"SQLITE_FILE": {
			Value: &configv1.SecretValue_PlainText{PlainText: ""},
		},
	})
	sqlite.SetCommandLineService(sqliteCmd)

	sqliteMcp := &configv1.McpUpstreamService{}
	sqliteMcp.SetToolAutoDiscovery(true)
	sqlite.SetMcpService(sqliteMcp)

	BuiltinTemplates = append(BuiltinTemplates, sqlite)

	// Everything
	everything := &configv1.UpstreamServiceConfig{}
	everything.SetId("everything")
	everything.SetName("Everything (Search)")
	everything.SetVersion("1.0.0")
	everything.SetTags([]string{"search", "windows", "fs"})
	everything.SetConfigurationSchema(`{
  "type": "object",
  "title": "Everything Configuration",
  "properties": {}
}`)

	everythingCmd := &configv1.CommandLineUpstreamService{}
	everythingCmd.SetCommand("npx -y @modelcontextprotocol/server-everything")
	everything.SetCommandLineService(everythingCmd)

	everythingMcp := &configv1.McpUpstreamService{}
	everythingMcp.SetToolAutoDiscovery(true)
	everything.SetMcpService(everythingMcp)

	BuiltinTemplates = append(BuiltinTemplates, everything)

	// Google Drive
	gdrive := &configv1.UpstreamServiceConfig{}
	gdrive.SetId("gdrive")
	gdrive.SetName("Google Drive")
	gdrive.SetVersion("1.0.0")
	gdrive.SetTags([]string{"storage", "google", "cloud"})
	gdrive.SetConfigurationSchema(`{
  "type": "object",
  "title": "Google Drive Configuration",
  "properties": {
      "Note": {
        "type": "string",
        "title": "Setup Required",
        "description": "This server requires OAuth setup which is complex via simple env vars. This is a placeholder for the official node server.",
        "readOnly": true
      }
  }
}`)

	gdriveCmd := &configv1.CommandLineUpstreamService{}
	gdriveCmd.SetCommand("npx -y @modelcontextprotocol/server-gdrive")
	gdrive.SetCommandLineService(gdriveCmd)

	gdriveMcp := &configv1.McpUpstreamService{}
	gdriveMcp.SetToolAutoDiscovery(true)
	gdrive.SetMcpService(gdriveMcp)

	BuiltinTemplates = append(BuiltinTemplates, gdrive)
}
