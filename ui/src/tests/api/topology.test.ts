/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi } from 'vitest';
import { GET } from '@/app/api/v1/topology/route';
import { NextResponse } from 'next/server';

// Mock dependencies
vi.mock('@/lib/server/mock-db', () => ({
    MockDB: {
        services: [
            { name: "test-service", version: "1.0.0", disable: false, http_service: { address: "http://test:8080" }, id: "srv-test-1" },
            { name: "disabled-service", version: "0.5.0", disable: true, id: "srv-test-2" }
        ]
    }
}));

describe('API: Topology', () => {
    it('returns a valid graph structure', async () => {
        const response = await GET();
        const json = await response.json();

        expect(json).toHaveProperty('core');
        expect(json).toHaveProperty('clients');
        expect(json.core.type).toBe('NODE_TYPE_CORE');
        expect(json.core.children).toBeDefined();

        // Check services
        const services = json.core.children.filter((c: any) => c.type === 'NODE_TYPE_SERVICE');
        expect(services).toHaveLength(2);

        const activeService = services.find((s: any) => s.label === 'test-service');
        expect(activeService.status).toBe('NODE_STATUS_ACTIVE');
        expect(activeService.metrics).toBeDefined();

        const disabledService = services.find((s: any) => s.label === 'disabled-service');
        expect(disabledService.status).toBe('NODE_STATUS_INACTIVE');
    });

    it('adds middleware nodes to core', async () => {
        const response = await GET();
        const json = await response.json();

        const middlewares = json.core.children.filter((c: any) => c.type === 'NODE_TYPE_MIDDLEWARE');
        expect(middlewares.length).toBeGreaterThan(0);
    });
});
