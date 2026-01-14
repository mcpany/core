/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from "@/lib/types";
import { Database, FileText, Github, Globe, Server, Terminal } from "lucide-react";

export interface ServiceTemplate {
  id: string;
  name: string;
  description: string;
  icon: any; // Lucide icon component
  config: Partial<UpstreamServiceConfig>;
  fields?: {
    name: string;
    label: string;
    placeholder: string;
    key: string; // Key path in config object (e.g. "httpService.address")
  }[];
}

export const SERVICE_TEMPLATES: ServiceTemplate[] = [
  {
    id: "empty",
    name: "Custom Service",
    description: "Configure a service from scratch.",
    icon: Server,
    config: {
      name: "",
      type: "http" as any, // Default to http
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
        command: "npx -y @modelcontextprotocol/server-postgres postgresql://user:password@localhost:5432/dbname",
        env: {},
      } as any, // Using 'as any' to bypass strict type check for partial config during template selection
    },
  },
  {
    id: "filesystem",
    name: "Filesystem",
    description: "Expose local files to the LLM.",
    icon: FileText,
    config: {
      name: "local-files",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-filesystem /path/to/directory",
        env: {},
      } as any,
    },
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
            "GITHUB_PERSONAL_ACCESS_TOKEN": "your-token-here"
        },
      } as any,
    },
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
            "BRAVE_API_KEY": "your-api-key"
        },
      } as any,
    },
  },
  {
    id: "sqlite",
    name: "SQLite",
    description: "Connect to a SQLite database.",
    icon: Database,
    config: {
      name: "sqlite-db",
      commandLineService: {
        command: "npx -y @modelcontextprotocol/server-sqlite /path/to/database.db",
        env: {},
      } as any,
    },
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
      } as any,
    },
  },
];
