/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */




/**
 * Configuration for running a marketplace tool.
 */
export interface MarketplaceItemConfig {
  /** The command to execute (e.g. "npx", "docker"). */
  command: string;
  /** Arguments to pass to the command. */
  args: string[];
  /** Environment variables required by the tool. */
  envVars: EnvVarDefinition[];
}

/**
 * Definition of an environment variable required by a marketplace tool.
 */
export interface EnvVarDefinition {
  /** The name of the environment variable (KEY). */
  name: string;
  /** A human-readable description of what this variable is for. */
  description: string;
  /** Whether this variable is mandatory. */
  required: boolean;
  /** The input type for the UI (text, password, or file path). */
  type: "text" | "password" | "path";
  /** If true, this value is also appended to the command args. */
  addToArgs?: boolean;
}

/**
 * Definition of a tool available in the marketplace.
 */
export interface MarketplaceItem {
  /** Unique identifier for the item. */
  id: string;
  /** Display name of the tool. */
  name: string;
  /** Brief description of the tool's capabilities. */
  description: string;
  /** Icon name (mapped to Lucide icons in UI). */
  icon: string;
  /** Default configuration for running the tool. */
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
