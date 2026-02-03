/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Configuration parameters for a marketplace item.
 * Defines how to execute the tool and what environment variables are needed.
 */
export interface MarketplaceItemConfig {
  /** The executable command (e.g., "npx", "docker"). */
  command: string;
  /** Arguments to pass to the command. */
  args: string[];
  /** Definitions of environment variables required by the item. */
  envVars: EnvVarDefinition[];
}

/**
 * Defines an environment variable required by a marketplace item.
 */
export interface EnvVarDefinition {
  /** The name of the environment variable (e.g., "API_KEY"). */
  name: string;
  /** A description of what this variable is for. */
  description: string;
  /** Whether this variable must be provided by the user. */
  required: boolean;
  /** The input type for the variable in the UI (text, password, path). */
  type: "text" | "password" | "path";
  /** If true, this value is also appended to the command args. */
  addToArgs?: boolean;
}

/**
 * Represents an item available in the marketplace.
 */
export interface MarketplaceItem {
  /** Unique identifier for the item. */
  id: string;
  /** Display name of the item. */
  name: string;
  /** Description of the item's functionality. */
  description: string;
  /** Name of the icon to display (mapped to Lucide icons). */
  icon: string;
  /** Configuration details for the item. */
  config: MarketplaceItemConfig;
}

/**
 * A registry of available marketplace items.
 *
 * This constant list defines the curated set of tools that users can easily install.
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
