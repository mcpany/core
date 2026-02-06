// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package marketplace

// DefaultRegistry is the embedded list of certified community servers.
// This replaces the fragile frontend parsing logic with a source of truth.
var DefaultRegistry = []CommunityServer{
	{
		Name:        "PostgreSQL",
		Description: "Read-only database access for PostgreSQL",
		Category:    "Database",
		URL:         "https://github.com/modelcontextprotocol/servers/tree/main/src/postgres",
		Tags:        []string{"database", "sql", "postgres"},
		Command:     "npx -y @modelcontextprotocol/server-postgres",
		ConfigurationSchema: `{
  "type": "object",
  "title": "PostgreSQL Configuration",
  "properties": {
    "args": {
        "type": "array",
        "title": "Arguments",
        "items": {
            "type": "string",
            "title": "Connection URL",
            "description": "postgresql://user:pass@host:5432/db"
        },
        "minItems": 1,
        "maxItems": 1
    }
  },
  "required": ["args"]
}`,
	},
	{
		Name:        "GitHub",
		Description: "Integration with GitHub API for repository management",
		Category:    "Developer Tools",
		URL:         "https://github.com/modelcontextprotocol/servers/tree/main/src/github",
		Tags:        []string{"git", "version-control", "github"},
		Command:     "npx -y @modelcontextprotocol/server-github",
		ConfigurationSchema: `{
  "type": "object",
  "title": "GitHub Configuration",
  "properties": {
    "GITHUB_PERSONAL_ACCESS_TOKEN": {
      "type": "string",
      "title": "Personal Access Token",
      "description": "GitHub PAT with repo permissions"
    }
  },
  "required": ["GITHUB_PERSONAL_ACCESS_TOKEN"]
}`,
	},
	{
		Name:        "SQLite",
		Description: "SQLite database interaction",
		Category:    "Database",
		URL:         "https://github.com/modelcontextprotocol/servers/tree/main/src/sqlite",
		Tags:        []string{"database", "sql", "sqlite", "local"},
		Command:     "npx -y @modelcontextprotocol/server-sqlite",
		ConfigurationSchema: `{
  "type": "object",
  "title": "SQLite Configuration",
  "properties": {
    "args": {
        "type": "array",
        "title": "Arguments",
        "items": {
            "type": "string",
            "title": "Database File Path",
            "description": "/path/to/database.db"
        },
        "minItems": 1,
        "maxItems": 1
    }
  },
  "required": ["args"]
}`,
	},
	{
		Name:        "Slack",
		Description: "Slack integration for channel management and messaging",
		Category:    "Communication",
		URL:         "https://github.com/modelcontextprotocol/servers/tree/main/src/slack",
		Tags:        []string{"chat", "messaging", "slack"},
		Command:     "npx -y @modelcontextprotocol/server-slack",
		ConfigurationSchema: `{
  "type": "object",
  "title": "Slack Configuration",
  "properties": {
    "SLACK_BOT_TOKEN": {
      "type": "string",
      "title": "Bot Token",
      "description": "xoxb-..."
    },
    "SLACK_TEAM_ID": {
      "type": "string",
      "title": "Team ID",
      "description": "T..."
    }
  },
  "required": ["SLACK_BOT_TOKEN", "SLACK_TEAM_ID"]
}`,
	},
	{
		Name:        "Cloudflare",
		Description: "Manage Cloudflare resources",
		Category:    "Cloud",
		URL:         "https://github.com/modelcontextprotocol/servers/tree/main/src/cloudflare",
		Tags:        []string{"cloud", "dns", "cdn"},
		Command:     "npx -y @modelcontextprotocol/server-cloudflare",
		ConfigurationSchema: `{
  "type": "object",
  "title": "Cloudflare Configuration",
  "properties": {
    "CLOUDFLARE_API_TOKEN": {
      "type": "string",
      "title": "API Token",
      "description": "Token with Account.Read permissions"
    },
    "CLOUDFLARE_ACCOUNT_ID": {
      "type": "string",
      "title": "Account ID",
      "description": "Your Cloudflare Account ID"
    }
  },
  "required": ["CLOUDFLARE_API_TOKEN", "CLOUDFLARE_ACCOUNT_ID"]
}`,
	},
	{
		Name:        "Sentry",
		Description: "Sentry integration for error tracking",
		Category:    "Monitoring",
		URL:         "https://github.com/modelcontextprotocol/servers/tree/main/src/sentry",
		Tags:        []string{"monitoring", "error-tracking", "sentry"},
		Command:     "npx -y @modelcontextprotocol/server-sentry",
		ConfigurationSchema: `{
  "type": "object",
  "title": "Sentry Configuration",
  "properties": {
    "SENTRY_AUTH_TOKEN": {
      "type": "string",
      "title": "Auth Token",
      "description": "User Auth Token"
    }
  },
  "required": ["SENTRY_AUTH_TOKEN"]
}`,
	},
	{
		Name:        "Google Drive",
		Description: "Access Google Drive files",
		Category:    "Productivity",
		URL:         "https://github.com/modelcontextprotocol/servers/tree/main/src/gdrive",
		Tags:        []string{"storage", "files", "google"},
		Command:     "npx -y @modelcontextprotocol/server-gdrive",
		ConfigurationSchema: `{
  "type": "object",
  "title": "Google Drive Configuration",
  "properties": {
    "GOOGLE_CLIENT_ID": {
      "type": "string",
      "title": "Client ID"
    },
    "GOOGLE_CLIENT_SECRET": {
      "type": "string",
      "title": "Client Secret"
    }
  },
  "required": ["GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET"]
}`,
	},
    {
		Name:        "FileSystem",
		Description: "Local filesystem access",
		Category:    "Core",
		URL:         "https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem",
		Tags:        []string{"files", "local", "core"},
		Command:     "npx -y @modelcontextprotocol/server-filesystem",
		ConfigurationSchema: `{
  "type": "object",
  "title": "Filesystem Configuration",
  "properties": {
    "args": {
        "type": "array",
        "title": "Allowed Paths",
        "items": {
            "type": "string",
            "title": "Path",
            "description": "/path/to/allow"
        },
        "minItems": 1
    }
  },
  "required": ["args"]
}`,
	},
}
