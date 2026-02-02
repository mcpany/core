/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Distributed Tracing', () => {
  test('should grouping requests by trace_id', async ({ request, baseURL }) => {
    // 1. Seed Backend directly (bypass UI)
    // We assume backend is running at http://localhost:50050 or defined via env
    const backendUrl = process.env.BACKEND_URL || 'http://localhost:50050';
    const traceId = '4bf92f3577b34da6a3ce929d0e0e4736';

    // We use a unique ID to avoid collision with previous tests
    const uniqueTraceId = traceId.substring(0, 30) + Math.floor(Math.random() * 90 + 10);

    // Send Request 1
    await request.get(`${backendUrl}/healthz`, {
      headers: {
        'traceparent': `00-${uniqueTraceId}-0000000000000001-01`
      }
    });

    // Send Request 2
    await request.get(`${backendUrl}/healthz`, {
      headers: {
        'traceparent': `00-${uniqueTraceId}-0000000000000002-01`
      }
    });

    // Allow some time for async processing if any (though debugger ring is immediate)
    // But API route fetches audit logs? No, fetches debug entries.

    // 2. Fetch from Next.js API
    const response = await request.get(`${baseURL}/api/traces`);
    expect(response.ok()).toBeTruthy();

    const traces = await response.json();
    const trace = traces.find((t: any) => t.id === uniqueTraceId);

    // 3. Verify
    expect(trace).toBeDefined();
    // Should be a single trace object for this ID
    expect(trace.id).toBe(uniqueTraceId);

    // Check structure
    // Since we sent 2 requests with same traceID but different parentIDs (that don't exist),
    // the logic should create a Synthetic Root containing both as children.
    expect(trace.rootSpan).toBeDefined();
    expect(trace.rootSpan.children).toBeDefined();
    expect(trace.rootSpan.children.length).toBeGreaterThanOrEqual(2);

    // Check span names
    const names = trace.rootSpan.children.map((c: any) => c.name);
    // Note: The Debugger middleware might capture "GET /healthz"
    expect(names.some((n: string) => n.includes('GET /healthz'))).toBeTruthy();
  });
});
