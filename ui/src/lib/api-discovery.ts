/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { DiscoveryServiceClientImpl, IndexStatus, GrpcWebImpl } from "@proto/api/v1/discovery_service";

// The GrpcWebImpl needs to be configured with the host
// We assume window.location.origin for browser usage, or fallback to localhost
const host = typeof window !== 'undefined' ? window.location.origin : 'http://localhost:50050';

const transport = new GrpcWebImpl(host, {
  debug: false,
});

const client = new DiscoveryServiceClientImpl(transport);

/**
 * Represents a single search result.
 */
export interface ToolSearchResult {
    toolName: string;
    serviceName: string;
    description: string;
    relevance: number;
}

/**
 * Provides access to the discovery API.
 */
export const discoveryApi = {
    /**
     * Search for tools using a query string.
     */
    async searchTools(query: string, limit: number = 20): Promise<ToolSearchResult[]> {
        try {
            const response = await client.SearchTools({
                query,
                limit,
            });

            // Note: Protobuf fields are typically camelCase in the generated JS object from GrpcWebImpl
            // But we need to check the generated interface.
            // The generated code interface SearchToolsResponse has `results: SearchResult[]`.
            // SearchResult has `tool: ToolDefinition | undefined`, `relevance: number`, `serviceName: string`.

            return (response.results || []).map(r => ({
                toolName: r.tool?.name || "Unknown",
                serviceName: r.serviceName,
                description: r.tool?.description || "",
                relevance: r.relevance
            }));
        } catch (e) {
            console.error("Failed to search tools", e);
            throw e;
        }
    },

    /**
     * Get the current status of the tool index.
     */
    async getIndexStatus(): Promise<IndexStatus> {
        try {
            const response = await client.GetIndexStatus({});
            return response;
        } catch (e) {
            console.error("Failed to get index status", e);
            throw e;
        }
    }
};
