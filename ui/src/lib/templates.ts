/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from "@/lib/types";
import { Database, FileText, Github, Globe, Server, Activity, Cloud, MessageSquare, Map, Clock, Zap, CheckCircle2 } from "lucide-react";

/**
 * A template for creating a new service configuration.
 */
export interface ServiceTemplate {
  /** Unique identifier for the template. */
  id: string;
  /** Display name of the template. */
  name: string;
  /** Description of what the template provides. */
  description: string;
  /** Icon component for the template. */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  icon: any; // Lucide icon component
  /** Partial configuration provided by the template. */
  config: Partial<UpstreamServiceConfig>;
  /** The category of the service. */
  category?: string;
  /** Whether the service is featured. */
  featured?: boolean;
  /**
   * Optional list of fields that need to be filled in by the user.
   */
  fields?: {
    /** The name of the field (internal identifier). */
    name: string;
    /** The label to display for the field. */
    label: string;
    /** Placeholder text for the input. */
    placeholder: string;
    /** Key path in the config object where the value should be set (e.g. "httpService.address"). */
    key: string;
    /**
     * If set, the value will not replace the entire key content but will substitute this token.
     * Useful for command line arguments.
     */
    replaceToken?: string;
    /** Default value for the field. */
    defaultValue?: string;
    /** Input type (text, password, etc). Defaults to text. */
    type?: string;
  }[];
}

/**
 * A list of built-in service templates.
 */
export const SERVICE_TEMPLATES: ServiceTemplate[] = [
  {
    id: "empty",
    name: "Custom Service",
    description: "Configure a service from scratch.",
    category: "Other",
    icon: Server,
    config: {
      name: "",
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      httpService: { address: "" } as any,
    },
  },
  {
    id: "wttrin",
    name: "Weather (wttr.in)",
    description: "Get real-time weather information via wttr.in.",
    category: "Web",
    featured: true,
    icon: Cloud,
    config: {
      name: "weather",
      httpService: {
        address: "https://wttr.in",
        tools: [
            {
                name: "get_weather",
                description: "Get the weather forecast for a location.",
                call_id: "get_weather_call"
            }
        ],
        calls: {
            "get_weather_call": {
                endpoint_path: "/{{location}}?format=j1",
                method: "HTTP_METHOD_GET",
                input_schema: {
                    type: "object",
                    properties: {
                        location: { type: "string", description: "City name" }
                    },
                    required: ["location"]
                }
            }
        }
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
  },
  {
    id: "google-maps",
    name: "Google Maps",
    description: "Geocoding, places, and routing.",
    category: "Web",
    icon: Map,
    config: {
      name: "google-maps",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-google-maps",
        env: {
            "GOOGLE_MAPS_API_KEY": ""
        },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
    fields: [
      {
        name: "apiKey",
        label: "Google Maps API Key",
        placeholder: "AIza...",
        key: "commandLineService.env.GOOGLE_MAPS_API_KEY",
        type: "password"
      }
    ]
  },
  {
    id: "slack",
    name: "Slack",
    description: "Interact with Slack channels and messages.",
    category: "Productivity",
    icon: MessageSquare,
    config: {
      name: "slack",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-slack",
        env: {
            "SLACK_BOT_TOKEN": "",
            "SLACK_TEAM_ID": ""
        },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
    fields: [
      {
        name: "botToken",
        label: "Slack Bot Token",
        placeholder: "xoxb-...",
        key: "commandLineService.env.SLACK_BOT_TOKEN",
        type: "password"
      },
      {
        name: "teamId",
        label: "Slack Team ID",
        placeholder: "T...",
        key: "commandLineService.env.SLACK_TEAM_ID"
      }
    ]
  },
  {
    id: "linear",
    name: "Linear",
    description: "Manage issues and projects in Linear.",
    category: "Productivity",
    icon: CheckCircle2,
    config: {
      name: "linear",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-linear",
        env: {
            "LINEAR_API_KEY": ""
        },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
    fields: [
      {
        name: "apiKey",
        label: "Linear API Key",
        placeholder: "lin_api_...",
        key: "commandLineService.env.LINEAR_API_KEY",
        type: "password"
      }
    ]
  },
  {
    id: "time",
    name: "Time",
    description: "Check the current time in any timezone.",
    category: "Utility",
    icon: Clock,
    config: {
      name: "time",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-time",
        env: {},
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
  },
  {
    id: "postgres",
    name: "PostgreSQL",
    description: "Connect to a PostgreSQL database.",
    category: "Database",
    icon: Database,
    config: {
      name: "postgres-db",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-postgres {{CONNECTION_STRING}}",
        env: {},
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
    fields: [
      {
        name: "connectionString",
        label: "PostgreSQL Connection String",
        placeholder: "postgresql://user:password@localhost:5432/dbname",
        key: "commandLineService.command",
        replaceToken: "{{CONNECTION_STRING}}",
      }
    ]
  },
  {
    id: "sqlite",
    name: "SQLite",
    description: "Connect to a SQLite database.",
    category: "Database",
    icon: Database,
    config: {
      name: "sqlite-db",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-sqlite {{DB_PATH}}",
        env: {},
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
    fields: [
      {
        name: "dbPath",
        label: "Database Path",
        placeholder: "/path/to/database.db",
        key: "commandLineService.command",
        replaceToken: "{{DB_PATH}}",
      }
    ]
  },
  {
    id: "filesystem",
    name: "Filesystem",
    description: "Expose local files to the LLM.",
    category: "System",
    icon: FileText,
    config: {
      name: "local-files",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-filesystem {{ALLOWED_DIRECTORIES}}",
        env: {},
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
    fields: [
      {
        name: "directories",
        label: "Allowed Directories",
        placeholder: "/path/to/folder1 /path/to/folder2",
        key: "commandLineService.command",
        replaceToken: "{{ALLOWED_DIRECTORIES}}",
      }
    ]
  },
  {
    id: "github",
    name: "GitHub",
    description: "Integration with GitHub API.",
    category: "Dev Tools",
    icon: Github,
    config: {
      name: "github-integration",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-github",
        env: {
            "GITHUB_PERSONAL_ACCESS_TOKEN": ""
        },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
    fields: [
      {
        name: "token",
        label: "GitHub Personal Access Token",
        placeholder: "ghp_...",
        key: "commandLineService.env.GITHUB_PERSONAL_ACCESS_TOKEN",
        type: "password"
      }
    ]
  },
  {
    id: "sentry",
    name: "Sentry",
    description: "Access Sentry issues and errors.",
    category: "Dev Tools",
    icon: Activity,
    config: {
      name: "sentry",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-sentry",
        env: {
            "SENTRY_AUTH_TOKEN": ""
        },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
    fields: [
      {
        name: "token",
        label: "Sentry Auth Token",
        placeholder: "Enter your Sentry Auth Token",
        key: "commandLineService.env.SENTRY_AUTH_TOKEN",
        type: "password"
      }
    ]
  },
  {
    id: "cloudflare",
    name: "Cloudflare",
    description: "Manage Cloudflare resources.",
    category: "Cloud",
    icon: Cloud,
    config: {
      name: "cloudflare",
      commandLineService: {
        command: "npx -y @cloudflare/mcp-server-cloudflare",
        env: {
            "CLOUDFLARE_API_TOKEN": "",
            "CLOUDFLARE_ACCOUNT_ID": ""
        },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
    fields: [
      {
        name: "apiToken",
        label: "Cloudflare API Token",
        placeholder: "Enter your Cloudflare API Token",
        key: "commandLineService.env.CLOUDFLARE_API_TOKEN",
        type: "password"
      },
      {
        name: "accountId",
        label: "Cloudflare Account ID",
        placeholder: "Enter your Cloudflare Account ID",
        key: "commandLineService.env.CLOUDFLARE_ACCOUNT_ID",
      }
    ]
  },
  {
    id: "web-search",
    name: "Brave Search",
    description: "Web search capabilities using Brave.",
    category: "Web",
    icon: Globe,
    config: {
      name: "brave-search",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-brave-search",
        env: {
            "BRAVE_API_KEY": ""
        },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
    fields: [
      {
        name: "apiKey",
        label: "Brave API Key",
        placeholder: "BSA...",
        key: "commandLineService.env.BRAVE_API_KEY",
        type: "password"
      }
    ]
  },
  {
    id: "puppeteer",
    name: "Puppeteer",
    description: "Browser automation and scraping.",
    category: "Web",
    icon: Globe,
    config: {
      name: "puppeteer-browser",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-puppeteer",
        env: {},
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
  },
  {
    id: "everything",
    name: "Everything",
    description: "Reference server with all MCP features.",
    category: "Utility",
    icon: Zap,
    config: {
      name: "everything",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-everything",
        env: {},
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
  },
];
