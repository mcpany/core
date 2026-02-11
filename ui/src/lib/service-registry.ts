/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Service Registry acts as a client-side database of known MCP servers.
 * It provides metadata like descriptions, installation commands, and configuration schemas.
 *
 * This allows the UI to offer "one-click" installation or configuration wizards for popular tools
 * without needing to query the backend for everything or rely on external internet access.
 */

interface RegistryItem {
    name: string;
    repo: string;
    description: string;
    command: string;
    configurationSchema: Record<string, unknown>;
}

/**
 * The static registry of known community MCP servers.
 */
export const SERVICE_REGISTRY: RegistryItem[] = [
    {
        name: "weather-service",
        repo: "github.com/example/weather-service",
        description: "A simple weather service for demos.",
        command: "python weather.py",
        configurationSchema: {
            type: "object",
            properties: {
                apiKey: { type: "string", description: "API Key for weather provider" }
            }
        }
    },
    // ... more items
];
