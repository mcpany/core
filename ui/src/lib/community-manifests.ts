/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Defines the structure for a community server manifest.
 * This provides "smart defaults" for servers that don't have an official schema.
 */
export interface CommunityManifest {
    /** The recommended command to run the server. */
    command: string;
    /** Known environment variables with descriptions. */
    env: Record<string, string>;
    /** A helpful tip for the user. */
    description?: string;
}

export const COMMUNITY_MANIFESTS: Record<string, CommunityManifest> = {
    // Git
    "mcp-server-git": {
        command: "npx -y @modelcontextprotocol/server-git",
        env: {
            // While git usually takes args, some wrappers use envs.
            // But the official one uses args. We can't easily injection args into command string safely without a parser.
            // However, we can hint the user.
            // Actually, let's assume the user might need these for some auth.
        },
        description: "Note: The Git server typically requires the repository path as an argument. Append it to the command."
    },
    "server-git": {
         command: "npx -y @modelcontextprotocol/server-git",
         env: {},
         description: "Note: The Git server typically requires the repository path as an argument. Append it to the command."
    },

    // Postgres
    "mcp-server-postgres": {
        command: "npx -y @modelcontextprotocol/server-postgres",
        env: {
            "POSTGRES_URL": "postgresql://user:password@localhost:5432/dbname"
        },
        description: "Provide the connection string to your PostgreSQL database."
    },
    "server-postgres": {
        command: "npx -y @modelcontextprotocol/server-postgres",
        env: {
            "POSTGRES_URL": "postgresql://user:password@localhost:5432/dbname"
        },
        description: "Provide the connection string to your PostgreSQL database."
    },

    // Slack
    "mcp-server-slack": {
        command: "npx -y @modelcontextprotocol/server-slack",
        env: {
            "SLACK_BOT_TOKEN": "xoxb-...",
            "SLACK_TEAM_ID": "T..."
        },
        description: "Requires a Slack App with Bot Token scopes."
    },

    // Google Drive
    "mcp-server-google-drive": {
        command: "npx -y @modelcontextprotocol/server-google-drive",
        env: {
            "GOOGLE_CLIENT_ID": "OAuth Client ID",
            "GOOGLE_CLIENT_SECRET": "OAuth Client Secret"
        },
        description: "You need to set up OAuth credentials in Google Cloud Console."
    },

    // Filesystem
    "mcp-server-filesystem": {
        command: "npx -y @modelcontextprotocol/server-filesystem",
        env: {},
        description: "Append the directories you want to expose to the command (e.g., /path/to/folder)."
    },

    // Memory
    "mcp-server-memory": {
        command: "npx -y @modelcontextprotocol/server-memory",
        env: {},
        description: "Knowledge graph memory server. No configuration required."
    },

    // AWS
    "mcp-server-aws": {
        command: "npx -y @modelcontextprotocol/server-aws",
        env: {
            "AWS_ACCESS_KEY_ID": "AKIA...",
            "AWS_SECRET_ACCESS_KEY": "Secret...",
            "AWS_REGION": "us-east-1"
        },
        description: "AWS SDK credentials."
    },

    // Sentry
    "mcp-server-sentry": {
        command: "npx -y @modelcontextprotocol/server-sentry",
        env: {
            "SENTRY_AUTH_TOKEN": "Auth Token"
        },
        description: "Sentry integration."
    },

    // Linear (for testing/demo)
    "linear": {
        command: "npx -y @modelcontextprotocol/server-linear",
        env: {
            "LINEAR_API_KEY": "lin_api_..."
        },
        description: "Linear integration requires an API key."
    }
};
