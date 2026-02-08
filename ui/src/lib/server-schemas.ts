/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { CommunityServer } from "./marketplace-service";
import { UpstreamServiceConfig } from "./client";

/**
 * ServerSchemaStrategy defines how to identify a server and inject its configuration schema.
 */
interface ServerSchemaStrategy {
    /**
     * Checks if this strategy applies to the given community server.
     * @param server The community server object.
     * @returns True if this strategy matches.
     */
    matches(server: CommunityServer): boolean;

    /**
     * Generates the configuration schema and command overrides for the server.
     * @param server The community server object.
     * @returns A partial UpstreamServiceConfig with schema and command updates.
     */
    enrich(server: CommunityServer): Partial<UpstreamServiceConfig>;
}

// --- Strategies ---

const CloudflareStrategy: ServerSchemaStrategy = {
    matches: (server) => {
        const repoMatch = server.url.match(/github\.com\/([^/]+)\/([^/]+)/);
        const repoName = repoMatch ? repoMatch[2] : "";
        return repoName === "mcp-server-cloudflare" || server.name.toLowerCase().includes("cloudflare");
    },
    enrich: (server) => {
        return {
            configurationSchema: JSON.stringify({
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
            })
        };
    }
};

const PostgresStrategy: ServerSchemaStrategy = {
    matches: (server) => {
        const repoMatch = server.url.match(/github\.com\/([^/]+)\/([^/]+)/);
        const repoName = repoMatch ? repoMatch[2] : "";
        return repoName === "server-postgres" || server.name.toLowerCase().includes("postgres");
    },
    enrich: (server) => {
        return {
            configurationSchema: JSON.stringify({
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
            })
        };
    }
};

const GitHubStrategy: ServerSchemaStrategy = {
    matches: (server) => {
        const repoMatch = server.url.match(/github\.com\/([^/]+)\/([^/]+)/);
        const repoName = repoMatch ? repoMatch[2] : "";
        return repoName === "server-github" || server.name.toLowerCase().includes("github");
    },
    enrich: (server) => {
        return {
             configurationSchema: JSON.stringify({
                type: "object",
                title: "GitHub Configuration",
                properties: {
                    "GITHUB_PERSONAL_ACCESS_TOKEN": {
                        type: "string",
                        title: "Personal Access Token",
                        description: "Your GitHub Personal Access Token (classic or fine-grained)",
                    }
                },
                required: ["GITHUB_PERSONAL_ACCESS_TOKEN"]
            })
        };
    }
};

// --- Registry ---

const STRATEGIES: ServerSchemaStrategy[] = [
    CloudflareStrategy,
    PostgresStrategy,
    GitHubStrategy,
];

/**
 * Enriches a Community Server configuration with smart defaults and schemas based on known patterns.
 * @param server The community server to enrich.
 * @returns An UpstreamServiceConfig populated with defaults and schemas.
 */
export function enrichCommunityServerConfig(server: CommunityServer): UpstreamServiceConfig {
    // 1. Determine base command
    const isPython = server.tags.some(t => t.includes('🐍'));
    let command = "";
    const repoMatch = server.url.match(/github\.com\/([^/]+)\/([^/]+)/);

    if (repoMatch) {
        const owner = repoMatch[1];
        const repo = repoMatch[2];
        if (isPython) {
           command = `uvx ${repo}`;
        } else {
           if (owner === 'modelcontextprotocol' && repo.startsWith('server-')) {
               command = `npx -y @modelcontextprotocol/${repo}`;
           } else {
               command = `npx -y ${repo}`; // fallback for other JS repos, usually doesn't work directly but good start
           }
        }
    } else {
        command = isPython ? "uvx package-name" : "npx -y package-name";
    }

    // 2. Base Config
    const config: UpstreamServiceConfig = {
        id: server.name.toLowerCase().replace(/[^a-z0-9-]/g, '-'),
        name: server.name,
        configurationSchema: "",
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

    // 3. Apply Strategies
    for (const strategy of STRATEGIES) {
        if (strategy.matches(server)) {
            const enrichment = strategy.enrich(server);
            Object.assign(config, enrichment);
            // If enrichment has commandLineService updates, merge deeply?
            // For now, simple object assign is enough as enrichment usually only adds schema or overrides specific top-level props.
            if (enrichment.commandLineService) {
                config.commandLineService = { ...config.commandLineService!, ...enrichment.commandLineService };
            }
            break; // First match wins
        }
    }

    return config;
}
