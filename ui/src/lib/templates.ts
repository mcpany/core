/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from "@/lib/types";
import { Database, FileText, Github, Globe, Server } from "lucide-react";

/**
 * A template for creating a new service configuration.
 */
export interface ServiceTemplate {
  /** Unique identifier for the template. */
  id: string;
  /** Display name of the template. */
  name: string;
  /** Description of what the template provides. */
  description: string;
  /** Icon component for the template. */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  icon: any; // Lucide icon component
  /** Partial configuration provided by the template. */
  config: Partial<UpstreamServiceConfig>;
  /**
   * Optional list of fields that need to be filled in by the user.
   */
  fields?: {
    /** The name of the field (internal identifier). */
    name: string;
    /** The label to display for the field. */
    label: string;
    /** Placeholder text for the input. */
    placeholder: string;
    /** Key path in the config object where the value should be set (e.g. "httpService.address"). */
    key: string;
    /**
     * If set, the value will not replace the entire key content but will substitute this token.
     * Useful for command line arguments.
     */
    replaceToken?: string;
    /** Default value for the field. */
    defaultValue?: string;
  }[];
}

/**
 * A list of built-in service templates.
 */
export const SERVICE_TEMPLATES: ServiceTemplate[] = [
  {
    id: "empty",
    name: "Custom Service",
    description: "Configure a service from scratch.",
    icon: Server,
    config: {
      name: "",
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      httpService: { address: "" } as any,
    },
  },
  {
    id: "postgres",
    name: "PostgreSQL",
    description: "Connect to a PostgreSQL database.",
    icon: Database,
    config: {
      name: "postgres-db",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-postgres {{CONNECTION_STRING}}",
        env: {},
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
    fields: [
      {
        name: "connectionString",
        label: "PostgreSQL Connection String",
        placeholder: "postgresql://user:password@localhost:5432/dbname",
        key: "commandLineService.command",
        replaceToken: "{{CONNECTION_STRING}}",
      }
    ]
  },
  {
    id: "filesystem",
    name: "Filesystem",
    description: "Expose local files to the LLM.",
    icon: FileText,
    config: {
      name: "local-files",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-filesystem {{ALLOWED_DIRECTORIES}}",
        env: {},
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
    fields: [
      {
        name: "directories",
        label: "Allowed Directories",
        placeholder: "/path/to/folder1 /path/to/folder2",
        key: "commandLineService.command",
        replaceToken: "{{ALLOWED_DIRECTORIES}}",
      }
    ]
  },
  {
    id: "github",
    name: "GitHub",
    description: "Integration with GitHub API.",
    icon: Github,
    config: {
      name: "github-integration",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-github",
        env: {
            "GITHUB_PERSONAL_ACCESS_TOKEN": ""
        },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
    fields: [
      {
        name: "token",
        label: "GitHub Personal Access Token",
        placeholder: "ghp_...",
        key: "commandLineService.env.GITHUB_PERSONAL_ACCESS_TOKEN",
      }
    ]
  },
  {
    id: "web-search",
    name: "Brave Search",
    description: "Web search capabilities using Brave.",
    icon: Globe,
    config: {
      name: "brave-search",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-brave-search",
        env: {
            "BRAVE_API_KEY": ""
        },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
    fields: [
      {
        name: "apiKey",
        label: "Brave API Key",
        placeholder: "BSA...",
        key: "commandLineService.env.BRAVE_API_KEY",
      }
    ]
  },
  {
    id: "sqlite",
    name: "SQLite",
    description: "Connect to a SQLite database.",
    icon: Database,
    config: {
      name: "sqlite-db",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-sqlite {{DB_PATH}}",
        env: {},
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
    fields: [
      {
        name: "dbPath",
        label: "Database Path",
        placeholder: "/path/to/database.db",
        key: "commandLineService.command",
        replaceToken: "{{DB_PATH}}",
      }
    ]
  },
  {
    id: "puppeteer",
    name: "Puppeteer",
    description: "Browser automation and scraping.",
    icon: Globe,
    config: {
      name: "puppeteer-browser",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-puppeteer",
        env: {},
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
    },
  },
];
