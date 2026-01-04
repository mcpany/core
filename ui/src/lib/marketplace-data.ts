/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

export interface MarketplaceItem {
    id: string;
    name: string;
    description: string;
    author: string;
    icon: string; // Lucide icon name
    category: 'database' | 'productivity' | 'filesystem' | 'development' | 'search' | 'other';
    config: {
        command: string;
        args: string[];
        envVars: {
            name: string;
            description: string;
            required: boolean;
            placeholder?: string;
            type?: 'text' | 'password' | 'path';
        }[];
    };
}

export const MARKETPLACE_ITEMS: MarketplaceItem[] = [
    {
        id: "filesystem",
        name: "Filesystem",
        description: "Access and manipulate files on the local system. Securely sandbox directory access.",
        author: "Model Context Protocol",
        icon: "FolderOpen",
        category: "filesystem",
        config: {
            command: "npx",
            args: ["-y", "@modelcontextprotocol/server-filesystem", "${ALLOWED_PATH}"],
            envVars: [
                {
                    name: "ALLOWED_PATH",
                    description: "The absolute path to the directory you want to expose.",
                    required: true,
                    placeholder: "/path/to/directory",
                    type: "path"
                }
            ]
        }
    },
    {
        id: "sqlite",
        name: "SQLite",
        description: "Query and manage SQLite databases directly. Perfect for local data analysis.",
        author: "Model Context Protocol",
        icon: "Database",
        category: "database",
        config: {
            command: "npx",
            args: ["-y", "@modelcontextprotocol/server-sqlite", "${DB_PATH}"],
            envVars: [
                {
                    name: "DB_PATH",
                    description: "Path to the SQLite database file.",
                    required: true,
                    placeholder: "/path/to/database.db",
                    type: "path"
                }
            ]
        }
    },
    {
        id: "github",
        name: "GitHub",
        description: "Interact with GitHub repositories, issues, and pull requests.",
        author: "Model Context Protocol",
        icon: "Github",
        category: "development",
        config: {
            command: "npx",
            args: ["-y", "@modelcontextprotocol/server-github"],
            envVars: [
                {
                    name: "GITHUB_PERSONAL_ACCESS_TOKEN",
                    description: "GitHub Personal Access Token (PAT) with repo permissions.",
                    required: true,
                    type: "password"
                }
            ]
        }
    },
    {
        id: "postgres",
        name: "PostgreSQL",
        description: "Full PostgreSQL database access with schema inspection and query capabilities.",
        author: "Model Context Protocol",
        icon: "DatabaseZap",
        category: "database",
        config: {
            command: "npx",
            args: ["-y", "@modelcontextprotocol/server-postgres", "${POSTGRES_URL}"],
            envVars: [
                {
                    name: "POSTGRES_URL",
                    description: "Connection string (e.g., postgres://user:pass@localhost:5432/db)",
                    required: true,
                    placeholder: "postgres://user:password@localhost:5432/dbname",
                    type: "password"
                }
            ]
        }
    },
    {
        id: "brave-search",
        name: "Brave Search",
        description: "Perform web searches using the privacy-focused Brave Search API.",
        author: "Model Context Protocol",
        icon: "Search",
        category: "search",
        config: {
            command: "npx",
            args: ["-y", "@modelcontextprotocol/server-brave-search"],
            envVars: [
                {
                    name: "BRAVE_API_KEY",
                    description: "Your Brave Search API Key.",
                    required: true,
                    type: "password"
                }
            ]
        }
    },
    {
        id: "google-drive",
        name: "Google Drive",
        description: "Access Google Drive files and folders. Requires OAuth credentials.",
        author: "Model Context Protocol",
        icon: "Cloud",
        category: "productivity",
        config: {
            command: "npx",
            args: ["-y", "@modelcontextprotocol/server-gdrive"],
            envVars: [
                {
                    name: "GOOGLE_CLIENT_ID",
                    description: "OAuth Client ID.",
                    required: true,
                    type: "text"
                },
                {
                    name: "GOOGLE_CLIENT_SECRET",
                    description: "OAuth Client Secret.",
                    required: true,
                    type: "password"
                }
            ]
        }
    },
    {
        id: "slack",
        name: "Slack",
        description: "Send messages and manage channels in Slack workspace.",
        author: "Model Context Protocol",
        icon: "MessageSquare",
        category: "productivity",
        config: {
            command: "npx",
            args: ["-y", "@modelcontextprotocol/server-slack"],
            envVars: [
                {
                    name: "SLACK_BOT_TOKEN",
                    description: "Slack Bot User OAuth Token (xoxb-...).",
                    required: true,
                    type: "password"
                },
                {
                    name: "SLACK_TEAM_ID",
                    description: "Slack Team ID (T...).",
                    required: true,
                    type: "text"
                }
            ]
        }
    },
    {
        id: "memory",
        name: "Memory",
        description: "Persistent knowledge graph memory for your AI assistant.",
        author: "Model Context Protocol",
        icon: "BrainCircuit",
        category: "other",
        config: {
            command: "npx",
            args: ["-y", "@modelcontextprotocol/server-memory"],
            envVars: []
        }
    },
    {
        id: "sequential-thinking",
        name: "Sequential Thinking",
        description: "A tool to help the model think through complex problems step-by-step.",
        author: "Model Context Protocol",
        icon: "ListOrdered",
        category: "other",
        config: {
            command: "npx",
            args: ["-y", "@modelcontextprotocol/server-sequential-thinking"],
            envVars: []
        }
    },
    {
        id: "puppeteer",
        name: "Puppeteer",
        description: "Headless browser automation for web scraping and testing.",
        author: "Model Context Protocol",
        icon: "Globe",
        category: "development",
        config: {
            command: "npx",
            args: ["-y", "@modelcontextprotocol/server-puppeteer"],
            envVars: []
        }
    }
];
