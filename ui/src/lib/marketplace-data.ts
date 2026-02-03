/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */




/**
 * Configuration parameters for a marketplace item.
 * Defines the command to run and the environment variables required.
 */
export interface MarketplaceItemConfig {
  /** The executable command (e.g., "npx", "docker"). */
  command: string;
  /** Arguments to pass to the command. */
  args: string[];
  /** Definitions of environment variables that the user needs to provide. */
  envVars: EnvVarDefinition[];
}

/**
 * Defines a required environment variable for a marketplace item.
 * Used to generate the input form for the user.
 */
export interface EnvVarDefinition {
  /** The name of the environment variable (e.g., "API_KEY"). */
  name: string;
  /** A user-friendly description of what this variable is for. */
  description: string;
  /** Whether this variable is required. */
  required: boolean;
  /** The input type for the UI (e.g., "password" masks the input). */
  type: "text" | "password" | "path";
  /** If true, this value is also appended to the command args (e.g., for some CLI tools). */
  addToArgs?: boolean;
}

/**
 * Represents an item available in the MCP Marketplace.
 */
export interface MarketplaceItem {
  /** Unique identifier for the item. */
  id: string;
  /** Display name of the item. */
  name: string;
  /** Short description of the item's functionality. */
  description: string;
  /** Name of the icon to display (mapped to Lucide icons). */
  icon: string;
  /** Configuration template for the item. */
  config: MarketplaceItemConfig;
}

/**
 * A curated list of available MCP tools for the marketplace.
 * This list is used to populate the "Add Service" wizard.
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
