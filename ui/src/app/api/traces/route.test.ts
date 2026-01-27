/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { GET } from './route';
import { NextResponse } from 'next/server';
import { vi, describe, it, expect, beforeEach } from 'vitest';

// Mock fetch
global.fetch = vi.fn();

describe('GET /api/traces', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('should fetch traces from backend and transform them', async () => {
        const mockEntries = [
            {
                id: '1',
                timestamp: '2023-01-01T00:00:00Z',
                method: 'POST',
                path: '/test',
                status: 200,
                duration: 10000000, // 10ms
                request_body: '{"foo":"bar"}',
                response_body: '{"baz":"qux"}'
            }
        ];

        (global.fetch as any).mockResolvedValue({
            ok: true,
            json: async () => mockEntries
        });

        const req = new Request('http://localhost/api/traces');
        const res = await GET(req);

        const json = await res.json();

        expect(json).toHaveLength(1);
        expect(json[0].id).toBe('1');
        expect(json[0].rootSpan.name).toBe('POST /test');
        expect(json[0].rootSpan.input).toEqual({ foo: 'bar' });
        expect(json[0].rootSpan.output).toEqual({ baz: 'qux' });
        expect(json[0].totalDuration).toBe(10);
    });

    it('should handle fetch error', async () => {
        (global.fetch as any).mockResolvedValue({
            ok: false,
            statusText: 'Internal Server Error'
        });

        const req = new Request('http://localhost/api/traces');
        const res = await GET(req);
        const json = await res.json();

        expect(json).toEqual([]);
    });
});
