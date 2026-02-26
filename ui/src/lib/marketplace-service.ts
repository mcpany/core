/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { apiClient, UpstreamServiceConfig } from "@/lib/client";

/**
 * A collection of services, typically organized by theme or use case.
 *
 * Summary: Definition of a service collection.
 *
 * Side Effects: None.
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
 *
 * Summary: Definition of an external marketplace.
 *
 * Side Effects: None.
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
 *
 * Summary: Definition of an external server.
 *
 * Side Effects: None.
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
 *
 * Summary: Definition of a community server.
 *
 * Side Effects: None.
 */
export interface CommunityServer {
    /** The category or section where this server was found (e.g., "Browser Automation"). */
    category: string;
    /** The name of the server. */
    name: string;
    /** The URL to the server's repository or documentation. */
    url: string;
    /** A brief description of the server's capabilities. */
    description: string;
    /** A list of tags or keywords associated with the server (e.g., emojis). */
    tags: string[];
}

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
 *
 * Summary: Service to manage marketplace interactions.
 *
 * Side Effects: None.
 */
export const marketplaceService = {
  /**
   * Fetches the official collections provided by MCP Any.
   *
   * Summary: Fetches official collections.
   * @returns A promise that resolves to a list of service collections.
   *
   * Side Effects: Makes a GET request to listCollections API.
   */
  fetchOfficialCollections: async (): Promise<ServiceCollection[]> => {
    try {
        const collections = await apiClient.listCollections();
        return collections as ServiceCollection[];
    } catch (e) {
        console.error("Failed to fetch official collections", e);
        return [];
    }
  },

  /**
   * Fetches the list of known public marketplaces.
   *
   * Summary: Fetches known public marketplaces.
   * @returns A promise that resolves to a list of external marketplaces.
   *
   * Side Effects: None.
   */
  fetchPublicMarketplaces: async (): Promise<ExternalMarketplace[]> => {
    return PUBLIC_MARKETPLACES;
  },

  /**
   * Fetches available servers from a specific external marketplace.
   *
   * Summary: Fetches servers from an external marketplace.
   * @param marketplaceId The ID of the marketplace to query.
   * @returns A promise that resolves to a list of external servers.
   *
   * Side Effects: Makes external network requests (mocked).
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
   * Fetches and parses the Awesome MCP Servers list from GitHub.
   *
   * Summary: Fetches community servers from GitHub.
   * @returns A promise that resolves to a list of CommunityServer objects.
   *
   * Side Effects: Makes a network request to GitHub.
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
   * Imports a collection from a URL.
   *
   * Summary: Imports a collection.
   * @param url The URL of the collection to import.
   * @returns A promise that resolves to the imported collection.
   *
   * Side Effects: Makes a network request.
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
   *
   * Summary: Fetches local collections.
   * @returns A list of locally stored service collections.
   *
   * Side Effects: Reads from localStorage.
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
   *
   * Summary: Saves a local collection.
   * @param collection The collection to save.
   *
   * Side Effects: Writes to localStorage.
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
   *
   * Summary: Deletes a local collection.
   * @param name The name of the collection to delete.
   *
   * Side Effects: Writes to localStorage.
   */
  deleteLocalCollection: (name: string) => {
      if (typeof window === 'undefined') return;
      const current = marketplaceService.fetchLocalCollections();
      const newCols = current.filter(c => c.name !== name);
      localStorage.setItem('mcp_local_collections', JSON.stringify(newCols));
  }
};
