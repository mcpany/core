// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

import { apiClient } from "./client";

/**
 * Definition of an indexed tool.
 */
export interface IndexedTool {
    name: string;
    description: string;
    category: string;
    installed: boolean;
    sourceUrl: string;
    tags: string[];
}

/**
 * Statistics about the index.
 */
export interface IndexStats {
    totalTools: number;
    totalSearches: number;
    hits: number;
    misses: number;
}

/**
 * Client for interacting with the Index Service (Lazy-MCP).
 */
export const indexClient = {
    /**
     * Searches the tool index.
     *
     * @param query - The search query.
     * @param page - The page number (1-based).
     * @param limit - The page size.
     * @returns A promise resolving to the search results and total count.
     */
    search: async (query: string, page = 1, limit = 20): Promise<{ tools: IndexedTool[], total: number }> => {
        // We reuse the fetchWithAuth logic from apiClient or just use fetch if we assume authenticated session.
        // Since `fetchWithAuth` is not exported from client.ts, we need to replicate it or modify client.ts to export it.
        // Or, better, add `searchIndex` to `apiClient` in `client.ts`.
        // But the plan was to create `index-client.ts`.
        // I'll assume I can use fetch and rely on the browser cookie or token handling,
        // OR I should have added it to `client.ts`.
        // Adding to `client.ts` is cleaner but `client.ts` is large.
        // Let's modify `client.ts` to export `fetchWithAuth`? No, it's local.
        // I will copy the fetch logic here for now, it's simple enough.

        const headers = new Headers();
        if (typeof window !== 'undefined') {
            const token = localStorage.getItem('mcp_auth_token');
            if (token) {
                headers.set('Authorization', `Basic ${token}`);
            }
        }

        const params = new URLSearchParams({
            query,
            page: page.toString(),
            limit: limit.toString()
        });

        const res = await fetch(`/api/v1/index/search?${params.toString()}`, { headers });
        if (!res.ok) throw new Error('Failed to search index');
        const data = await res.json();

        // Map snake_case to camelCase if needed, but our Go backend uses proto JSON mapping which handles camelCase by default for proto3 json.
        // But I used `IndexedTool` in proto.
        // Let's verify JSON output of proto.
        return {
            tools: (data.tools || []).map((t: any) => ({
                name: t.name,
                description: t.description,
                category: t.category,
                installed: t.installed,
                sourceUrl: t.source_url || t.sourceUrl,
                tags: t.tags || []
            })),
            total: data.total || 0
        };
    },

    /**
     * Seeds the index (for testing).
     *
     * @param tools - The tools to seed.
     * @param clear - Whether to clear existing index.
     */
    seed: async (tools: IndexedTool[], clear = false): Promise<number> => {
        const headers = new Headers({ 'Content-Type': 'application/json' });
        if (typeof window !== 'undefined') {
            const token = localStorage.getItem('mcp_auth_token');
            if (token) {
                headers.set('Authorization', `Basic ${token}`);
            }
        }

        const payload = {
            tools: tools.map(t => ({
                name: t.name,
                description: t.description,
                category: t.category,
                installed: t.installed,
                source_url: t.sourceUrl,
                tags: t.tags
            })),
            clear
        };

        const res = await fetch('/api/v1/index/seed', {
            method: 'POST',
            headers,
            body: JSON.stringify(payload)
        });

        if (!res.ok) throw new Error('Failed to seed index');
        const data = await res.json();
        return data.count;
    },

    /**
     * Gets index statistics.
     */
    getStats: async (): Promise<IndexStats> => {
        const headers = new Headers();
        if (typeof window !== 'undefined') {
            const token = localStorage.getItem('mcp_auth_token');
            if (token) {
                headers.set('Authorization', `Basic ${token}`);
            }
        }

        const res = await fetch('/api/v1/index/stats', { headers });
        if (!res.ok) throw new Error('Failed to get stats');
        const data = await res.json();

        return {
            totalTools: data.total_tools || data.totalTools || 0,
            totalSearches: data.total_searches || data.totalSearches || 0,
            hits: data.hits || 0,
            misses: data.misses || 0
        };
    }
};
