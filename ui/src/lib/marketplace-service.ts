/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig, apiClient, CommunityServer } from "@/lib/client";

export type { CommunityServer };

/**
 * A collection of services, typically organized by theme or use case.
 */
export interface ServiceCollection {
  /** The name of the collection. */
  name: string;
  /** A description of what the collection provides. */
  description: string;
  /** The author or maintainer of the collection. */
  author: string;
  /** The version of the collection. */
  version: string;
  /** The list of service configurations included in the collection. */
  services: UpstreamServiceConfig[];
}

/**
 * An external marketplace where MCP servers can be discovered.
 */
export interface ExternalMarketplace {
  /** Unique identifier for the marketplace. */
  id: string;
  /** Display name of the marketplace. */
  name: string;
  /** URL of the marketplace website or API. */
  url: string;
  /** Description of the marketplace. */
  description: string;
  /** Icon name or URL for the marketplace. */
  icon?: string;
}

/**
 * A server listed in an external marketplace.
 */
export interface ExternalServer {
  /** Unique identifier for the server. */
  id: string;
  /** Display name of the server. */
  name: string;
  /** Description of the server. */
  description: string;
  /** The author of the server. */
  author?: string;
  /** The configuration to install this server. */
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
                env: { "DB_PATH": { plainText: "./data.db", validationRegex: "" } }, // placeholders
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

            autoDiscoverTool: false,
            configError: "",
            configurationSchema: "",
            readOnly: false,
            tags: []
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

/**
 * Service for interacting with internal and external marketplaces.
 */
export const marketplaceService = {
  /**
   * Fetches the official collections provided by MCP Any.
   * @returns A promise that resolves to a list of service collections.
   */
  fetchOfficialCollections: async (): Promise<ServiceCollection[]> => {
    // In future: fetch('https://raw.githubusercontent.com/mcpany/marketplace/main/collections.json')
    return new Promise(resolve => setTimeout(() => resolve(MOCK_OFFICIAL_COLLECTIONS), 500));
  },

  /**
   * Fetches the list of known public marketplaces.
   * @returns A promise that resolves to a list of external marketplaces.
   */
  fetchPublicMarketplaces: async (): Promise<ExternalMarketplace[]> => {
    return PUBLIC_MARKETPLACES;
  },

  /**
   * Fetches available servers from a specific external marketplace.
   * @param marketplaceId The ID of the marketplace to query.
   * @returns A promise that resolves to a list of external servers.
   */
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
                        env: { "LINEAR_API_KEY": { plainText: "", validationRegex: "" } },
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

                    autoDiscoverTool: false,
                    configError: "",
                    configurationSchema: "",
                    readOnly: false,
                    tags: []
                }

            }
        ];
    }
    return [];
  },

  /**
   * Fetches the certified list of community servers from the backend registry.
   * @returns A promise that resolves to a list of CommunityServer objects.
   */
  fetchCommunityServers: async (): Promise<CommunityServer[]> => {
      try {
          return await apiClient.getCommunityServers();
      } catch (error) {
          console.error("Error fetching community servers:", error);
          return [];
      }
  },


  /**
   * Imports a collection from a URL.
   * @param url The URL of the collection to import.
   * @returns A promise that resolves to the imported collection.
   */
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
  },

  // Local Storage Logic

  /**
   * Fetches collections stored locally in the browser.
   * @returns A list of locally stored service collections.
   */
  fetchLocalCollections: (): ServiceCollection[] => {
      if (typeof window === 'undefined') return [];
      try {
          const stored = localStorage.getItem('mcp_local_collections');
          return stored ? JSON.parse(stored) : [];
      } catch (e) {
          console.error("Failed to parse local collections", e);
          return [];
      }
  },

  /**
   * Saves a collection to local storage.
   * @param collection The collection to save.
   */
  saveLocalCollection: (collection: ServiceCollection) => {
      if (typeof window === 'undefined') return;
      const current = marketplaceService.fetchLocalCollections();
      // Update if exists or append
      const idx = current.findIndex(c => c.name === collection.name); // Simple dedupe by name for now
      if (idx >= 0) {
          current[idx] = collection;
      } else {
          current.push(collection);
      }
      localStorage.setItem('mcp_local_collections', JSON.stringify(current));
  },

  /**
   * Deletes a locally stored collection.
   * @param name The name of the collection to delete.
   */
  deleteLocalCollection: (name: string) => {
      if (typeof window === 'undefined') return;
      const current = marketplaceService.fetchLocalCollections();
      const newCols = current.filter(c => c.name !== name);
      localStorage.setItem('mcp_local_collections', JSON.stringify(newCols));
  }
};
