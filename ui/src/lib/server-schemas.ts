/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Returns the JSON Schema configuration string for a given repository.
 * @param owner The repository owner (e.g. "modelcontextprotocol").
 * @param repo The repository name (e.g. "server-github").
 * @returns The JSON Schema string or null if not found.
 */
export function getSchemaForServer(owner: string, repo: string): string | null {
  const key = `${owner.toLowerCase()}/${repo.toLowerCase()}`;

  // Check strict matches
  if (SERVER_SCHEMAS[key]) {
    return JSON.stringify(SERVER_SCHEMAS[key]);
  }

  // Fallback: Check if repo name matches a known service (e.g. "mcp-server-cloudflare" -> matches "cloudflare")
  // This is useful for forks or different naming conventions
  const normalizedRepo = repo.toLowerCase().replace(/^mcp-server-|^server-/, '');

  // Define generic mappings for common services regardless of owner
  const GENERIC_SCHEMAS: Record<string, any> = {
    'cloudflare': SERVER_SCHEMAS['modelcontextprotocol/server-cloudflare'] || SERVER_SCHEMAS['cloudflare/mcp-server-cloudflare'],
    'postgres': SERVER_SCHEMAS['modelcontextprotocol/server-postgres'],
    'postgresql': SERVER_SCHEMAS['modelcontextprotocol/server-postgres'],
    'github': SERVER_SCHEMAS['modelcontextprotocol/server-github'],
    'gitlab': SERVER_SCHEMAS['modelcontextprotocol/server-gitlab'],
    'slack': SERVER_SCHEMAS['modelcontextprotocol/server-slack'],
    'google-maps': SERVER_SCHEMAS['modelcontextprotocol/server-google-maps'],
    'brave-search': SERVER_SCHEMAS['modelcontextprotocol/server-brave-search'],
    'sentry': SERVER_SCHEMAS['modelcontextprotocol/server-sentry'],
    'linear': SERVER_SCHEMAS['modelcontextprotocol/server-linear'],
  };

  if (GENERIC_SCHEMAS[normalizedRepo]) {
      return JSON.stringify(GENERIC_SCHEMAS[normalizedRepo]);
  }

  return null;
}

const SERVER_SCHEMAS: Record<string, any> = {
  // Cloudflare
  "cloudflare/mcp-server-cloudflare": {
    type: "object",
    title: "Cloudflare Configuration",
    properties: {
      "CLOUDFLARE_API_TOKEN": {
        type: "string",
        title: "Cloudflare API Token",
        description: "Your Cloudflare API Token (Account.Read permissions required)",
      },
      "CLOUDFLARE_ACCOUNT_ID": {
        type: "string",
        title: "Account ID",
        description: "Your Cloudflare Account ID"
      }
    },
    required: ["CLOUDFLARE_API_TOKEN", "CLOUDFLARE_ACCOUNT_ID"]
  },
  "modelcontextprotocol/server-cloudflare": {
      type: "object",
      title: "Cloudflare Configuration",
      properties: {
        "CLOUDFLARE_API_TOKEN": {
          type: "string",
          title: "Cloudflare API Token",
          description: "Your Cloudflare API Token (Account.Read permissions required)",
        },
        "CLOUDFLARE_ACCOUNT_ID": {
          type: "string",
          title: "Account ID",
          description: "Your Cloudflare Account ID"
        }
      },
      required: ["CLOUDFLARE_API_TOKEN", "CLOUDFLARE_ACCOUNT_ID"]
  },

  // Postgres
  "modelcontextprotocol/server-postgres": {
    type: "object",
    title: "PostgreSQL Configuration",
    properties: {
      "POSTGRES_URL": {
        type: "string",
        title: "Connection URL",
        description: "postgresql://user:pass@host:5432/db",
      }
    },
    required: ["POSTGRES_URL"]
  },

  // GitHub
  "modelcontextprotocol/server-github": {
    type: "object",
    title: "GitHub Configuration",
    properties: {
      "GITHUB_PERSONAL_ACCESS_TOKEN": {
        type: "string",
        title: "Personal Access Token",
        description: "Your GitHub Personal Access Token with repo permissions.",
      }
    },
    required: ["GITHUB_PERSONAL_ACCESS_TOKEN"]
  },

  // GitLab
  "modelcontextprotocol/server-gitlab": {
    type: "object",
    title: "GitLab Configuration",
    properties: {
      "GITLAB_PERSONAL_ACCESS_TOKEN": {
        type: "string",
        title: "Personal Access Token",
        description: "Your GitLab Personal Access Token.",
      },
      "GITLAB_API_URL": {
        type: "string",
        title: "API URL (Optional)",
        description: "Defaults to https://gitlab.com/api/v4",
        default: "https://gitlab.com/api/v4"
      }
    },
    required: ["GITLAB_PERSONAL_ACCESS_TOKEN"]
  },

  // Slack
  "modelcontextprotocol/server-slack": {
    type: "object",
    title: "Slack Configuration",
    properties: {
      "SLACK_BOT_TOKEN": {
        type: "string",
        title: "Bot Token",
        description: "Your Slack Bot User OAuth Token (xoxb-...)",
      },
      "SLACK_TEAM_ID": {
        type: "string",
        title: "Team ID",
        description: "Your Slack Team ID (e.g. T12345678)",
      }
    },
    required: ["SLACK_BOT_TOKEN", "SLACK_TEAM_ID"]
  },

  // Google Maps
  "modelcontextprotocol/server-google-maps": {
    type: "object",
    title: "Google Maps Configuration",
    properties: {
      "GOOGLE_MAPS_API_KEY": {
        type: "string",
        title: "API Key",
        description: "Your Google Maps API Key.",
      }
    },
    required: ["GOOGLE_MAPS_API_KEY"]
  },

  // Brave Search
  "modelcontextprotocol/server-brave-search": {
    type: "object",
    title: "Brave Search Configuration",
    properties: {
      "BRAVE_API_KEY": {
        type: "string",
        title: "API Key",
        description: "Your Brave Search API Key.",
      }
    },
    required: ["BRAVE_API_KEY"]
  },

  // Sentry
  "modelcontextprotocol/server-sentry": {
    type: "object",
    title: "Sentry Configuration",
    properties: {
      "SENTRY_AUTH_TOKEN": {
        type: "string",
        title: "Auth Token",
        description: "Your Sentry Auth Token.",
      }
    },
    required: ["SENTRY_AUTH_TOKEN"]
  },

  // Linear
  "modelcontextprotocol/server-linear": {
    type: "object",
    title: "Linear Configuration",
    properties: {
      "LINEAR_API_KEY": {
        type: "string",
        title: "API Key",
        description: "Your Linear API Key.",
      }
    },
    required: ["LINEAR_API_KEY"]
  }
};
