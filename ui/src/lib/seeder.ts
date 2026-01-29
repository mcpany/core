/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * SeedTool represents a tool to be seeded.
 */
export interface SeedTool {
    name: string;
    description: string;
    serviceId: string;
    inputSchema: Record<string, any>;
    output: any;
}

/**
 * SeedRequest represents the payload for seeding the server state.
 */
export interface SeedRequest {
    services: any[];
    tools: SeedTool[];
    traffic: any[];
}

/**
 * Helper to construct seed data.
 */
export const createSeedData = (services: any[] = [], tools: SeedTool[] = [], traffic: any[] = []): SeedRequest => ({
      services,
      tools,
      traffic
});
