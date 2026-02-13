/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

export interface ServiceRegistryItem {
  id: string;
  name: string;
  repo: string; // Used for matching against community URLs
  command: string;
  description: string;
  configurationSchema: Record<string, any>;
}

/**
 * Registry of available services for the marketplace.
 * Maps service identifiers to their configuration and deployment details.
 */
export const SERVICE_REGISTRY: ServiceRegistryItem[] = [
  // --- Databases ---
  {
    id: "postgres",
    name: "PostgreSQL",
    repo: "modelcontextprotocol/servers/tree/main/src/postgres",
    command: "npx -y @modelcontextprotocol/server-postgres",
    description: "Read-only access to PostgreSQL databases",
    configurationSchema: {
      type: "object",
      required: ["POSTGRES_URL"],
      properties: {
        POSTGRES_URL: {
          type: "string",
          title: "Connection URL",
          description: "postgresql://user:password@host:5432/db",
          format: "password"
        }
      }
    }
  },
  {
    id: "sqlite",
    name: "SQLite",
    repo: "modelcontextprotocol/servers/tree/main/src/sqlite",
    command: "npx -y @modelcontextprotocol/server-sqlite",
    description: "Access and query SQLite databases",
    configurationSchema: {
      type: "object",
      required: ["DB_PATH"],
      properties: {
        DB_PATH: {
          type: "string",
          title: "Database Path",
          description: "Absolute path to the SQLite file (e.g. /data/mydb.sqlite)",
          default: "mcp.db"
        }
      }
    }
  },

  // --- Cloud Platforms ---
  {
    id: "cloudflare",
    name: "Cloudflare",
    repo: "modelcontextprotocol/servers/tree/main/src/cloudflare",
    command: "npx -y @modelcontextprotocol/server-cloudflare",
    description: "Manage Cloudflare resources",
    configurationSchema: {
      type: "object",
      required: ["CLOUDFLARE_API_TOKEN", "CLOUDFLARE_ACCOUNT_ID"],
      properties: {
        CLOUDFLARE_API_TOKEN: {
          type: "string",
          title: "API Token",
          description: "Cloudflare API Token with Account.Read permissions",
          format: "password"
        },
        CLOUDFLARE_ACCOUNT_ID: {
          type: "string",
          title: "Account ID",
          description: "Your Cloudflare Account ID"
        }
      }
    }
  },
  {
    id: "aws-kb",
    name: "AWS Knowledge Base",
    repo: "modelcontextprotocol/servers/tree/main/src/aws-kb",
    command: "npx -y @modelcontextprotocol/server-aws-kb",
    description: "Interact with AWS Bedrock Knowledge Bases",
    configurationSchema: {
      type: "object",
      required: ["AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION"],
      properties: {
        AWS_ACCESS_KEY_ID: {
          type: "string",
          title: "Access Key ID",
          description: "AWS Access Key ID",
          format: "password"
        },
        AWS_SECRET_ACCESS_KEY: {
          type: "string",
          title: "Secret Access Key",
          description: "AWS Secret Access Key",
          format: "password"
        },
        AWS_REGION: {
          type: "string",
          title: "Region",
          description: "AWS Region (e.g. us-east-1)",
          default: "us-east-1"
        }
      }
    }
  },

  // --- Developer Tools ---
  {
    id: "github",
    name: "GitHub",
    repo: "modelcontextprotocol/servers/tree/main/src/github",
    command: "npx -y @modelcontextprotocol/server-github",
    description: "Access GitHub repositories, issues, and PRs",
    configurationSchema: {
      type: "object",
      required: ["GITHUB_PERSONAL_ACCESS_TOKEN"],
      properties: {
        GITHUB_PERSONAL_ACCESS_TOKEN: {
          type: "string",
          title: "Personal Access Token",
          description: "GitHub PAT with repo scope",
          format: "password"
        }
      }
    }
  },
  {
    id: "gitlab",
    name: "GitLab",
    repo: "modelcontextprotocol/servers/tree/main/src/gitlab",
    command: "npx -y @modelcontextprotocol/server-gitlab",
    description: "Access GitLab projects and merge requests",
    configurationSchema: {
      type: "object",
      required: ["GITLAB_PERSONAL_ACCESS_TOKEN", "GITLAB_API_URL"],
      properties: {
        GITLAB_PERSONAL_ACCESS_TOKEN: {
          type: "string",
          title: "Personal Access Token",
          description: "GitLab PAT with api scope",
          format: "password"
        },
        GITLAB_API_URL: {
          type: "string",
          title: "GitLab API URL",
          description: "Base URL for GitLab API",
          default: "https://gitlab.com/api/v4"
        }
      }
    }
  },
  {
    id: "sentry",
    name: "Sentry",
    repo: "modelcontextprotocol/servers/tree/main/src/sentry",
    command: "npx -y @modelcontextprotocol/server-sentry",
    description: "Retrieve issues and events from Sentry",
    configurationSchema: {
      type: "object",
      required: ["SENTRY_AUTH_TOKEN"],
      properties: {
        SENTRY_AUTH_TOKEN: {
          type: "string",
          title: "Auth Token",
          description: "Sentry Auth Token",
          format: "password"
        }
      }
    }
  },

  // --- Productivity ---
  {
    id: "slack",
    name: "Slack",
    repo: "modelcontextprotocol/servers/tree/main/src/slack",
    command: "npx -y @modelcontextprotocol/server-slack",
    description: "Send messages and read channels in Slack",
    configurationSchema: {
      type: "object",
      required: ["SLACK_BOT_TOKEN", "SLACK_TEAM_ID"],
      properties: {
        SLACK_BOT_TOKEN: {
          type: "string",
          title: "Bot User OAuth Token",
          description: "xoxb-...",
          format: "password"
        },
        SLACK_TEAM_ID: {
          type: "string",
          title: "Team ID",
          description: "Workspace Team ID (e.g. T01234567)"
        }
      }
    }
  },
  {
    id: "linear",
    name: "Linear",
    repo: "jerhadf/linear-mcp-server",
    command: "npx -y linear-mcp-server",
    description: "Manage issues and cycles in Linear",
    configurationSchema: {
      type: "object",
      required: ["LINEAR_API_KEY"],
      properties: {
        LINEAR_API_KEY: {
          type: "string",
          title: "API Key",
          description: "Linear Personal API Key",
          format: "password"
        }
      }
    }
  },
  {
    id: "google-drive",
    name: "Google Drive",
    repo: "modelcontextprotocol/servers/tree/main/src/gdrive",
    command: "npx -y @modelcontextprotocol/server-gdrive",
    description: "Access and search Google Drive files",
    configurationSchema: {
      type: "object",
      required: ["GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET"],
      properties: {
        GOOGLE_CLIENT_ID: {
          type: "string",
          title: "Client ID",
          description: "Google Cloud Client ID"
        },
        GOOGLE_CLIENT_SECRET: {
          type: "string",
          title: "Client Secret",
          description: "Google Cloud Client Secret",
          format: "password"
        }
      }
    }
  },

  // --- Web & Search ---
  {
    id: "brave-search",
    name: "Brave Search",
    repo: "modelcontextprotocol/servers/tree/main/src/brave-search",
    command: "npx -y @modelcontextprotocol/server-brave-search",
    description: "Perform web searches using Brave Search API",
    configurationSchema: {
      type: "object",
      required: ["BRAVE_API_KEY"],
      properties: {
        BRAVE_API_KEY: {
          type: "string",
          title: "API Key",
          description: "Brave Search API Key",
          format: "password"
        }
      }
    }
  },
  {
    id: "google-maps",
    name: "Google Maps",
    repo: "modelcontextprotocol/servers/tree/main/src/google-maps",
    command: "npx -y @modelcontextprotocol/server-google-maps",
    description: "Access Google Maps Platform APIs",
    configurationSchema: {
      type: "object",
      required: ["GOOGLE_MAPS_API_KEY"],
      properties: {
        GOOGLE_MAPS_API_KEY: {
          type: "string",
          title: "API Key",
          description: "Google Maps API Key",
          format: "password"
        }
      }
    }
  },
  {
    id: "fetch",
    name: "Fetch",
    repo: "modelcontextprotocol/servers/tree/main/src/fetch",
    command: "npx -y @modelcontextprotocol/server-fetch",
    description: "Fetch web content as markdown",
    configurationSchema: {
      type: "object",
      required: [],
      properties: {}
    }
  },
  {
    id: "puppeteer",
    name: "Puppeteer",
    repo: "modelcontextprotocol/servers/tree/main/src/puppeteer",
    command: "npx -y @modelcontextprotocol/server-puppeteer",
    description: "Control a headless browser for automation",
    configurationSchema: {
      type: "object",
      required: [],
      properties: {}
    }
  },

  // --- Local System ---
  {
    id: "filesystem",
    name: "FileSystem",
    repo: "modelcontextprotocol/servers/tree/main/src/filesystem",
    command: "npx -y @modelcontextprotocol/server-filesystem",
    description: "Access local files and directories",
    configurationSchema: {
      type: "object",
      required: ["ALLOWED_DIRECTORIES"],
      properties: {
        ALLOWED_DIRECTORIES: {
          type: "string",
          title: "Allowed Directories",
          description: "Comma-separated list of absolute paths to allow access to",
          default: "/tmp"
        }
      }
    }
  },
  {
    id: "memory",
    name: "Memory",
    repo: "modelcontextprotocol/servers/tree/main/src/memory",
    command: "npx -y @modelcontextprotocol/server-memory",
    description: "Persistent knowledge graph memory",
    configurationSchema: {
      type: "object",
      required: [],
      properties: {}
    }
  },
  {
    id: "time",
    name: "Time",
    repo: "modelcontextprotocol/servers/tree/main/src/time",
    command: "npx -y @modelcontextprotocol/server-time",
    description: "Get current time and timezone information",
    configurationSchema: {
      type: "object",
      required: [],
      properties: {}
    }
  },

  // --- External Integrations ---
  {
    id: "everart",
    name: "EverArt",
    repo: "modelcontextprotocol/servers/tree/main/src/everart",
    command: "npx -y @modelcontextprotocol/server-everart",
    description: "Generate images using EverArt models",
    configurationSchema: {
      type: "object",
      required: ["EVERART_API_KEY", "EVERART_MEMBER_ID"],
      properties: {
        EVERART_API_KEY: {
          type: "string",
          title: "API Key",
          description: "EverArt API Key",
          format: "password"
        },
        EVERART_MEMBER_ID: {
          type: "string",
          title: "Member ID",
          description: "EverArt Member ID"
        }
      }
    }
  },
  {
    id: "sequential-thinking",
    name: "Sequential Thinking",
    repo: "modelcontextprotocol/servers/tree/main/src/sequentialthinking",
    command: "npx -y @modelcontextprotocol/server-sequentialthinking",
    description: "Tool for dynamic step-by-step problem solving",
    configurationSchema: {
      type: "object",
      required: [],
      properties: {}
    }
  }
];
