/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Configuration schema for running a marketplace item.
 * Defines the executable command, arguments, and environment variables required.
 */
export interface MarketplaceItemConfig {
  /** The command to execute (e.g., "npx", "docker"). */
  command: string;
  /** List of static arguments to pass to the command. */
  args: string[];
  /** List of environment variable definitions required by the service. */
  envVars: EnvVarDefinition[];
}

/**
 * Defines an environment variable required by a marketplace service.
 */
export interface EnvVarDefinition {
  /** The name of the environment variable (e.g., "API_KEY"). */
  name: string;
  /** A human-readable description of what this variable is for. */
  description: string;
  /** Whether this variable must be provided by the user. */
  required: boolean;
  /** The type of input field to render in the UI. */
  type: "text" | "password" | "path";
  /** If true, the value of this variable is appended to the command arguments instead of just being set in the environment. */
  addToArgs?: boolean;
}

/**
 * Represents a service available in the marketplace.
 * Contains metadata and configuration instructions.
 */
export interface MarketplaceItem {
  /** Unique identifier for the marketplace item. */
  id: string;
  /** Display name of the service. */
  name: string;
  /** Short description of what the service does. */
  description: string;
  /** Name of the Lucide icon to display (mapped in UI). */
  icon: string; // We'll map string to Lucide icon in the UI
  /** Configuration details for deploying this service. */
  config: MarketplaceItemConfig;
}

/**
 * A curated list of available marketplace items.
 * These are the default services that users can easily install.
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
