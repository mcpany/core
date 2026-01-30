/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from "@/lib/client";

/**
 * RegistryItem defines how to identify and configure a known MCP server.
 */
export interface RegistryItem {
    /** Unique ID for the registry item. */
    id: string;
    /** Friendly name. */
    name: string;
    /** Regex to match the repository URL or name. */
    matchRegex: RegExp;
    /** JSON Schema for the configuration form. */
    schema: any;
    /**
     * Configures the service based on the schema values.
     * @param config The base configuration (with command already set from repo).
     * @param values The values from the schema form.
     * @returns The modified configuration.
     */
    configure: (config: UpstreamServiceConfig, values: Record<string, string>) => UpstreamServiceConfig;
}

const REGISTRY: RegistryItem[] = [
    {
        id: "cloudflare",
        name: "Cloudflare",
        matchRegex: /mcp-server-cloudflare|cloudflare/i,
        schema: {
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
        configure: (config, values) => {
            if (!config.commandLineService) return config;
            config.commandLineService.env = {
                ...config.commandLineService.env,
                "CLOUDFLARE_API_TOKEN": { plainText: values["CLOUDFLARE_API_TOKEN"] },
                "CLOUDFLARE_ACCOUNT_ID": { plainText: values["CLOUDFLARE_ACCOUNT_ID"] }
            };
            return config;
        }
    },
    {
        id: "postgres",
        name: "PostgreSQL",
        matchRegex: /server-postgres|postgres/i,
        schema: {
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
        configure: (config, values) => {
            if (!config.commandLineService) return config;
            // Postgres server usually takes the connection string as an argument
            config.commandLineService.command = `${config.commandLineService.command} ${values["POSTGRES_URL"]}`;
            return config;
        }
    },
    {
        id: "github",
        name: "GitHub",
        matchRegex: /server-github|github\.com\/modelcontextprotocol\/server-github/i,
        schema: {
            type: "object",
            title: "GitHub Configuration",
            properties: {
                "GITHUB_PERSONAL_ACCESS_TOKEN": {
                    type: "string",
                    title: "Personal Access Token",
                    description: "A GitHub PAT with appropriate permissions.",
                }
            },
            required: ["GITHUB_PERSONAL_ACCESS_TOKEN"]
        },
        configure: (config, values) => {
            if (!config.commandLineService) return config;
            config.commandLineService.env = {
                ...config.commandLineService.env,
                "GITHUB_PERSONAL_ACCESS_TOKEN": { plainText: values["GITHUB_PERSONAL_ACCESS_TOKEN"] }
            };
            return config;
        }
    },
    {
        id: "slack",
        name: "Slack",
        matchRegex: /server-slack|slack/i,
        schema: {
            type: "object",
            title: "Slack Configuration",
            properties: {
                "SLACK_BOT_TOKEN": {
                    type: "string",
                    title: "Bot Token",
                    description: "xoxb-...",
                },
                "SLACK_TEAM_ID": {
                    type: "string",
                    title: "Team ID",
                    description: "T...",
                }
            },
            required: ["SLACK_BOT_TOKEN", "SLACK_TEAM_ID"]
        },
        configure: (config, values) => {
            if (!config.commandLineService) return config;
            config.commandLineService.env = {
                ...config.commandLineService.env,
                "SLACK_BOT_TOKEN": { plainText: values["SLACK_BOT_TOKEN"] },
                "SLACK_TEAM_ID": { plainText: values["SLACK_TEAM_ID"] }
            };
            return config;
        }
    },
    {
        id: "filesystem",
        name: "Filesystem",
        matchRegex: /server-filesystem|filesystem/i,
        schema: {
            type: "object",
            title: "Filesystem Configuration",
            properties: {
                "ALLOWED_PATHS": {
                    type: "string",
                    title: "Allowed Paths",
                    description: "Space-separated list of directories to expose (e.g. /home/user/docs)",
                }
            },
            required: ["ALLOWED_PATHS"]
        },
        configure: (config, values) => {
            if (!config.commandLineService) return config;
            // Append paths to command
            config.commandLineService.command = `${config.commandLineService.command} ${values["ALLOWED_PATHS"]}`;
            return config;
        }
    },
    {
        id: "sentry",
        name: "Sentry",
        matchRegex: /server-sentry|sentry/i,
        schema: {
            type: "object",
            title: "Sentry Configuration",
            properties: {
                "SENTRY_AUTH_TOKEN": {
                    type: "string",
                    title: "Auth Token",
                    description: "Sentry Authentication Token",
                }
            },
            required: ["SENTRY_AUTH_TOKEN"]
        },
        configure: (config, values) => {
            if (!config.commandLineService) return config;
            config.commandLineService.env = {
                ...config.commandLineService.env,
                "SENTRY_AUTH_TOKEN": { plainText: values["SENTRY_AUTH_TOKEN"] }
            };
            return config;
        }
    },
     {
        id: "google-maps",
        name: "Google Maps",
        matchRegex: /server-google-maps|google-maps/i,
        schema: {
            type: "object",
            title: "Google Maps Configuration",
            properties: {
                "GOOGLE_MAPS_API_KEY": {
                    type: "string",
                    title: "API Key",
                    description: "Google Maps API Key",
                }
            },
            required: ["GOOGLE_MAPS_API_KEY"]
        },
        configure: (config, values) => {
            if (!config.commandLineService) return config;
            config.commandLineService.env = {
                ...config.commandLineService.env,
                "GOOGLE_MAPS_API_KEY": { plainText: values["GOOGLE_MAPS_API_KEY"] }
            };
            return config;
        }
    }
];

/**
 * Resolves a server configuration based on the repository URL or name.
 * @param repoOrName The repository URL or name.
 * @returns The matching RegistryItem or undefined.
 */
export function resolveServerConfig(repoOrName: string): RegistryItem | undefined {
    return REGISTRY.find(item => item.matchRegex.test(repoOrName));
}

/**
 * Gets a registry item by its ID.
 * @param id The registry ID.
 * @returns The matching RegistryItem or undefined.
 */
export function getRegistryItemById(id: string): RegistryItem | undefined {
    return REGISTRY.find(item => item.id === id);
}
