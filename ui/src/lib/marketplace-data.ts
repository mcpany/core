/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */




/**
 * MarketplaceItemConfig description.
 *
 */
export interface MarketplaceItemConfig {
  command: string;
  args: string[];
  envVars: EnvVarDefinition[];
}

/**
 * EnvVarDefinition type definition.
 */
export interface EnvVarDefinition {
  name: string;
  description: string;
  required: boolean;
  type: "text" | "password" | "path";
  // If true, this value is also appended to the command args
  addToArgs?: boolean;
}

/**
 * MarketplaceItem type definition.
 */
export interface MarketplaceItem {
  id: string;
  name: string;
  description: string;
  icon: string; // We'll map string to Lucide icon in the UI
  config: MarketplaceItemConfig;
}

/**
 * The MARKETPLACE_ITEMS const.
 */
export const MARKETPLACE_ITEMS: MarketplaceItem[] = [
  {
    id: "filesystem",
    name: "Filesystem",
    description: "Read/Write access to local files and directories.",
    icon: "HardDrive",
    config: {
      command: "npx",
      args: ["-y", "@modelcontextprotocol/server-filesystem"],
      envVars: [
        {
          name: "ALLOWED_PATH",
          description: "Path to expose (e.g. /home/user/projects)",
          required: true,
          type: "path",
          addToArgs: true,
        },
      ],
    },
  },
  {
    id: "github",
    name: "GitHub",
    description: "Access GitHub repositories, issues, and pull requests.",
    icon: "Github",
    config: {
      command: "npx",
      args: ["-y", "@modelcontextprotocol/server-github"],
      envVars: [
        {
          name: "GITHUB_PERSONAL_ACCESS_TOKEN",
          description: "Personal Access Token (classic or fine-grained)",
          required: true,
          type: "password",
        },
      ],
    },
  },
  {
    id: "sqlite",
    name: "SQLite",
    description: "Query and manage SQLite databases.",
    icon: "Database",
    config: {
      command: "npx",
      args: ["-y", "@modelcontextprotocol/server-sqlite"],
      envVars: [
        {
          name: "DB_PATH",
          description: "Path to SQLite database file",
          required: true,
          type: "path",
          addToArgs: true,
        },
      ],
    },
  },
];
