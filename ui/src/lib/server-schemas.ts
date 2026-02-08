/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { CommunityServer } from "@/lib/marketplace-service";
import { UpstreamServiceConfig } from "@/lib/client";

export interface ServerSchemaStrategy {
    /**
     * pattern: A regex to match against the repo name or server name to identify the SERVICE TYPE.
     */
    pattern: RegExp;
    /**
     * schema: The JSON schema for the configuration.
     */
    schema: any;
}

const STRATEGIES: ServerSchemaStrategy[] = [
    {
        pattern: /cloudflare/i,
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
        }
    },
    {
        pattern: /postgres/i,
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
        }
    },
    {
        pattern: /github/i,
        schema: {
            type: "object",
            title: "GitHub Configuration",
            properties: {
                "GITHUB_PERSONAL_ACCESS_TOKEN": {
                    type: "string",
                    title: "Personal Access Token",
                    description: "GitHub PAT with repo permissions",
                }
            },
            required: ["GITHUB_PERSONAL_ACCESS_TOKEN"]
        }
    },
    {
        pattern: /slack/i,
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
        }
    },
    {
        pattern: /google-drive|gdrive/i,
        schema: {
            type: "object",
            title: "Google Drive Configuration",
            properties: {
                "GOOGLE_CLIENT_ID": {
                    type: "string",
                    title: "Client ID",
                },
                "GOOGLE_CLIENT_SECRET": {
                    type: "string",
                    title: "Client Secret",
                }
            },
            required: ["GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET"]
        }
    },
    {
        pattern: /aws|amazon/i,
        schema: {
            type: "object",
            title: "AWS Configuration",
            properties: {
                "AWS_ACCESS_KEY_ID": {
                    type: "string",
                    title: "Access Key ID",
                },
                "AWS_SECRET_ACCESS_KEY": {
                    type: "string",
                    title: "Secret Access Key",
                },
                "AWS_REGION": {
                    type: "string",
                    title: "Region",
                    default: "us-east-1"
                }
            },
            required: ["AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"]
        }
    },
    {
        pattern: /filesystem|fs/i,
        schema: {
            type: "object",
            title: "Filesystem Configuration",
            properties: {
                "ALLOWED_DIRECTORIES": {
                    type: "string",
                    title: "Allowed Directories",
                    description: "Space separated list of directories to expose (args)",
                    default: "."
                }
            },
            required: ["ALLOWED_DIRECTORIES"]
        }
    },
    {
        pattern: /brave-search|brave/i,
        schema: {
            type: "object",
            title: "Brave Search Configuration",
            properties: {
                "BRAVE_API_KEY": {
                    type: "string",
                    title: "Brave API Key",
                }
            },
            required: ["BRAVE_API_KEY"]
        }
    },
    {
        pattern: /fetch/i,
        schema: {
            type: "object",
            title: "Fetch Configuration",
            properties: {
                "FETCH_API_KEY": {
                    type: "string",
                    title: "API Key (Optional)",
                    description: "Optional API key for authenticated requests"
                }
            },
            required: []
        }
    },
    {
        pattern: /memory/i,
        schema: {
            type: "object",
            title: "Memory Configuration",
            properties: {},
            required: []
        }
    }
];

export function enrichCommunityServerConfig(server: CommunityServer): UpstreamServiceConfig {
    const isPython = server.tags.some(t => t.includes('🐍') || t.includes('python'));
    const repoMatch = server.url.match(/github\.com\/([^/]+)\/([^/]+)/);
    const repoInfo = repoMatch ? { owner: repoMatch[1], repo: repoMatch[2] } : null;

    // 1. Determine Command based on Repo URL (Truth)
    let command = "";
    if (repoInfo) {
        if (isPython) {
            command = `uvx ${repoInfo.repo}`;
        } else {
            if (repoInfo.owner === 'modelcontextprotocol' && repoInfo.repo.startsWith('server-')) {
                command = `npx -y @modelcontextprotocol/${repoInfo.repo}`;
            } else {
                command = `npx -y ${repoInfo.repo}`;
            }
        }
    } else {
         // Fallback Heuristic based on name
         command = isPython ? `uvx ${server.name.toLowerCase().replace(/ /g, '-')}` : `npx -y ${server.name.toLowerCase().replace(/ /g, '-')}`;
    }

    // 2. Determine Schema based on Service Identity (Name or Repo Name)
    // We match against server name OR repo name to guess "What kind of service is this?"
    const strategy = STRATEGIES.find(s => s.pattern.test(server.name) || (repoInfo && s.pattern.test(repoInfo.repo)));
    const schema = strategy ? JSON.stringify(strategy.schema) : "";

    return {
        id: server.name.toLowerCase().replace(/[^a-z0-9-]/g, '-'),
        name: server.name,
        configurationSchema: schema,
        version: "1.0.0",
        commandLineService: {
            command: command,
            env: {},
            workingDirectory: "",
            tools: [],
            resources: [],
            prompts: [],
            calls: {},
            communicationProtocol: 0,
            local: false
        },
        disable: false,
        sanitizedName: server.name.toLowerCase().replace(/[^a-z0-9-]/g, '-'),
        priority: 0,
        loadBalancingStrategy: 0,
        callPolicies: [],
        preCallHooks: [],
        postCallHooks: [],
        prompts: [],
        autoDiscoverTool: true,
        configError: "",
        tags: server.tags,
        readOnly: false
    };
}
