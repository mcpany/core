/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from "@/lib/client";

export interface ServiceCollection {
  name: string;
  description: string;
  author: string;
  version: string;
  services: UpstreamServiceConfig[];
}

export interface ExternalMarketplace {
  id: string;
  name: string;
  url: string;
  description: string;
  icon?: string;
}

export interface ExternalServer {
  id: string;
  name: string;
  description: string;
  author?: string;
  config: UpstreamServiceConfig; // Mapped to our config format
}

// Mock Data for "Official" Collections until we have the repo up
const MOCK_OFFICIAL_COLLECTIONS: ServiceCollection[] = [
  {
    name: "Data Engineering Stack",
    description: "Essential tools for data pipelines (PostgreSQL, Filesystem, Python)",
    author: "MCP Any Team",
    version: "1.0.0",
    services: [
        {
            id: "sqlite-db",
            name: "SQLite Database",
            version: "1.0.0",
            commandLineService: {
                command: "npx -y @modelcontextprotocol/server-sqlite",
                env: { "DB_PATH": { plainText: "./data.db" } }, // placeholders
                workingDirectory: "",
                tools: [],
                resources: [],
                prompts: [],
                calls: {},
                communicationProtocol: 0,
                local: false
            },
            disable: false,
            sanitizedName: "sqlite-db",
            priority: 0,
            loadBalancingStrategy: 0,
            callPolicies: [],
            preCallHooks: [],
            postCallHooks: [],
            prompts: [],
            autoDiscoverTool: false
        }
    ]
  },
  {
    name: "Web Dev Assistant",
    description: "GitHub, Browser, and Terminal tools for web development.",
    author: "MCP Any Team",
    version: "1.0.0",
    services: []
  }
];

const PUBLIC_MARKETPLACES: ExternalMarketplace[] = [
  {
    id: "mcpmarket",
    name: "MCP Market",
    url: "https://mcpmarket.com", // We might use a proxy or API if available
    description: "Community curated MCP servers.",
    icon: "Globe"
  },
  {
      id: "smithery",
      name: "Smithery",
      url: "https://smithery.ai",
      description: "Discover and manage AI agents and tools.",
      icon: "Box"
  }
];

export const marketplaceService = {
  fetchOfficialCollections: async (): Promise<ServiceCollection[]> => {
    // In future: fetch('https://raw.githubusercontent.com/mcpany/marketplace/main/collections.json')
    return new Promise(resolve => setTimeout(() => resolve(MOCK_OFFICIAL_COLLECTIONS), 500));
  },

  fetchPublicMarketplaces: async (): Promise<ExternalMarketplace[]> => {
    return PUBLIC_MARKETPLACES;
  },

  fetchExternalServers: async (marketplaceId: string): Promise<ExternalServer[]> => {
    // Mock fetching from external source
    // Real implementation would scrape or use API of the target marketplace
    if (marketplaceId === 'mcpmarket') {
        return [
            {
                id: 'linear',
                name: 'Linear',
                description: 'Linear issue tracking integration',
                author: 'Figma',
                config: {
                    id: 'linear',
                    name: 'Linear',
                    version: '1.0.0',
                    commandLineService: {
                        command: 'npx -y @modelcontextprotocol/server-linear',
                        env: { "LINEAR_API_KEY": { plainText: "" } },
                        workingDirectory: "",
                        tools: [],
                        resources: [],
                        prompts: [],
                        calls: {},
                        communicationProtocol: 0,
                        local: false
                    },
                    disable: false,
                    sanitizedName: "linear",
                    priority: 0,
                    loadBalancingStrategy: 0,
                    callPolicies: [],
                    preCallHooks: [],
                    postCallHooks: [],
                    prompts: [],
                    autoDiscoverTool: false
                }
            }
        ];
    }
    return [];
  },

  importCollection: async (url: string): Promise<ServiceCollection> => {
     // Fetch from URL, validate, return
     // Mock for now
     return {
         name: "Imported Collection",
         description: `Imported from ${url}`,
         author: "Unknown",
         version: "0.0.1",
         services: []
     };
  }
};
