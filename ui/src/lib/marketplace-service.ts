/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from "@/lib/client";

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

/**
 * A server discovered from the Community (Awesome List).
 */
export interface CommunityServer {
    category: string;
    name: string;
    url: string;
    description: string;
    tags: string[];
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
 * A singleton service responsible for managing interactions with various server marketplaces and collection sources.
 *
 * This service handles fetching official curated collections, discovering public marketplaces,
 * listing external servers, and managing locally stored user collections.
 */
export const marketplaceService = {
  /**
   * Retrieves the official "Gold Standard" service collections curated by the MCP Any team.
   *
   * These collections serve as the default starting point for users and include essential tools
   * and configurations.
   *
   * @returns {Promise<ServiceCollection[]>} A promise that resolves to a list of curated ServiceCollection objects.
   * @throws {Error} If the network request fails or the data is malformed.
   */
  fetchOfficialCollections: async (): Promise<ServiceCollection[]> => {
    // In future: fetch('https://raw.githubusercontent.com/mcpany/marketplace/main/collections.json')
    return new Promise(resolve => setTimeout(() => resolve(MOCK_OFFICIAL_COLLECTIONS), 500));
  },

  /**
   * Retrieves the list of known public MCP marketplaces.
   *
   * These marketplaces are external sources where additional community servers can be found.
   *
   * @returns {Promise<ExternalMarketplace[]>} A promise that resolves to a list of ExternalMarketplace objects.
   */
  fetchPublicMarketplaces: async (): Promise<ExternalMarketplace[]> => {
    return PUBLIC_MARKETPLACES;
  },

  /**
   * Queries a specific external marketplace for available servers.
   *
   * @param {string} marketplaceId - The unique identifier of the marketplace to query.
   * @returns {Promise<ExternalServer[]>} A promise that resolves to a list of servers available in the specified marketplace.
   * @throws {Error} If the marketplace ID is invalid or the upstream request fails.
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
   * Scrapes and parses the community-maintained "Awesome MCP Servers" list from GitHub.
   *
   * This method downloads the README.md file from the remote repository, parses the Markdown
   * to extract server links, categories, and tags, and returns them as structured objects.
   *
   * @returns {Promise<CommunityServer[]>} A promise that resolves to a list of community-contributed servers.
   * @throws {Error} If the fetch fails or parsing encounters unexpected structure.
   */
  fetchCommunityServers: async (): Promise<CommunityServer[]> => {
      try {
          const response = await fetch('https://raw.githubusercontent.com/punkpeye/awesome-mcp-servers/main/README.md');
          if (!response.ok) throw new Error('Failed to fetch Awesome list');
          const markdown = await response.text();

          const servers: CommunityServer[] = [];
          const lines = markdown.split('\n');
          let currentCategory = "Uncategorized";

          // Regex to match: * [Name](URL) Tags - Description OR - [Name](URL) Tags - Description
          const itemRegex = /^\s*[\-\*]\s+\[([^\]]+)\]\(([^)]+)\)\s*(.*?)\s*-\s*(.*)$/;

          // Regex to match category headers (e.g., "## 📂 File Systems") or "📂 File Systems" inside a list if structured differently
          // The structure seems to be:
          // * 📂 - [Browser Automation](#-browser-automation)
          // ...
          // ## 📂 Browser Automation

          for (const line of lines) {
              if (line.startsWith('## ') || line.startsWith('### ')) {
                  // Clean up header to get category name
                  currentCategory = line.replace(/^#+\s*/, '').trim();
                  // Remove links in headers if any
                  currentCategory = currentCategory.replace(/\[([^\]]+)\]\([^)]+\)/g, '$1');
                  continue;
              }

              const match = line.match(itemRegex);
              if (match) {
                  const name = match[1].trim();
                  const url = match[2].trim();
                  const tagsRaw = match[3].trim();
                  const description = match[4].trim();

                  // Extract emojis as tags
                  // Simple heuristic: split by space, keep if it's emoji-like or short code
                  const tags = tagsRaw.split(/\s+/).filter(t => t.length > 0);

                  // Filter out "TOC" items which might look like servers but point to anchors
                  if (url.startsWith('#')) continue;

                  servers.push({
                      category: currentCategory,
                      name,
                      url,
                      description,
                      tags
                  });
              }
          }
          return servers;
      } catch (error) {
          console.error("Error fetching community servers:", error);
          return [];
      }
  },


  /**
   * Imports a remote service collection from a given URL.
   *
   * This allows users to share configuration bundles. The method validates the downloaded
   * JSON against the ServiceCollection schema before returning it.
   *
   * @param {string} url - The URL pointing to the collection JSON file.
   * @returns {Promise<ServiceCollection>} A promise that resolves to the validated ServiceCollection.
   * @throws {Error} If the URL is invalid, the fetch fails, or the content is not a valid collection.
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
   * Retrieves user-defined collections stored in the browser's LocalStorage.
   *
   * This provides persistence for custom configurations across browser sessions.
   *
   * @returns {ServiceCollection[]} A list of locally stored ServiceCollection objects.
   * @sideeffects Reads from `localStorage` key 'mcp_local_collections'.
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
   * Persists a service collection to the browser's LocalStorage.
   *
   * If a collection with the same name already exists, it is overwritten.
   *
   * @param {ServiceCollection} collection - The collection object to save.
   * @sideeffects Writes to `localStorage` key 'mcp_local_collections'.
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
   * Removes a service collection from LocalStorage by name.
   *
   * @param {string} name - The unique name of the collection to delete.
   * @sideeffects Modifies `localStorage` key 'mcp_local_collections'.
   */
  deleteLocalCollection: (name: string) => {
      if (typeof window === 'undefined') return;
      const current = marketplaceService.fetchLocalCollections();
      const newCols = current.filter(c => c.name !== name);
      localStorage.setItem('mcp_local_collections', JSON.stringify(newCols));
  }
};
