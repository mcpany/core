/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * EnvVarDefinition defines the schema for an environment variable required by a marketplace item.
 *
 * @property name - The name of the environment variable (e.g. API_KEY).
 * @property description - A helpful description of what this variable is for.
 * @property required - Whether this variable must be provided.
 * @property type - The input type for the variable (text, password, path).
 * @property addToArgs - If true, the value is appended to the command arguments.
 */
export interface EnvVarDefinition {
  name: string;
  description: string;
  required: boolean;
  type: "text" | "password" | "path";
  addToArgs?: boolean;
}

/**
 * MarketplaceItemConfig defines the execution configuration for a marketplace item.
 *
 * @property command - The command to execute (e.g. npx).
 * @property args - The arguments to pass to the command.
 * @property envVars - The list of environment variables required.
 */
export interface MarketplaceItemConfig {
  command: string;
  args: string[];
  envVars: EnvVarDefinition[];
}

/**
 * MarketplaceItem represents a tool or service available in the marketplace.
 *
 * @property id - Unique identifier for the item.
 * @property name - Display name of the item.
 * @property description - Brief description of the item's functionality.
 * @property icon - Name of the Lucide icon to display.
 * @property config - The configuration required to run the item.
 */
export interface MarketplaceItem {
  id: string;
  name: string;
  description: string;
  icon: string; // We'll map string to Lucide icon in the UI
  config: MarketplaceItemConfig;
}

/**
 * MARKETPLACE_ITEMS defines the curated list of available marketplace items.
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
