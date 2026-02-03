/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import crypto from 'crypto';

// Ensure we use 127.0.0.1 to avoid IPv6 issues in CI, even if env var says localhost
let BACKEND_URL = process.env.BACKEND_URL || 'http://127.0.0.1:50050';
if (BACKEND_URL.includes('localhost')) {
    BACKEND_URL = BACKEND_URL.replace('localhost', '127.0.0.1');
}

test.describe('Distributed Tracing', () => {
  test('should visualize nested traces from W3C headers', async ({ page, request }) => {
    // Generate W3C compatible IDs
    // TraceID: 32 hex characters (16 bytes)
    // SpanID: 16 hex characters (8 bytes)
    const traceId = crypto.randomBytes(16).toString('hex');
    const parentSpanId = crypto.randomBytes(8).toString('hex');

    console.log(`Seeding trace: ${traceId} to ${BACKEND_URL}`);

    // Send two requests with the same Trace ID
    // This simulates a distributed trace where multiple spans belong to the same trace.
    // Both point to the same parent (simulating siblings).
    const headers = {
        'traceparent': `00-${traceId}-${parentSpanId}-01`,
        // Add fake user agent to be easily identifiable if needed, though traceId is enough
        'User-Agent': 'playwright-tracing-test'
    };

    // Retry sending requests if connection fails (e.g. backend starting up)
    await expect(async () => {
        const res = await request.get(`${BACKEND_URL}/health`, { headers });
        expect(res.status()).toBeGreaterThanOrEqual(200);
    }).toPass({ timeout: 10000 });

    await request.get(`${BACKEND_URL}/health`, { headers });

    // Poll for the trace to be available in the backend debugger API
    // This ensures data is present before we try to load the UI
    await expect(async () => {
        const res = await request.get(`${BACKEND_URL}/debug/entries`);
        expect(res.ok()).toBeTruthy();
        const entries = await res.json();
        // Check if any entry has our traceId (field might be trace_id or id depending on version/parsing)
        // We look for trace_id in the JSON
        const found = entries.some((e: any) => e.trace_id === traceId || e.id === traceId);
        expect(found).toBeTruthy();
    }).toPass({ timeout: 15000, intervals: [1000] });

    await page.goto('/traces');

    // Search for the trace to filter the list
    const searchInput = page.getByPlaceholder('Filter traces...');
    await expect(searchInput).toBeVisible();
    await searchInput.fill(traceId);

    // Wait for list to update and show the trace item
    // We expect the trace ID to appear in the list.
    // Use a more specific selector to avoid matching the input value itself if it remains visible
    const traceItem = page.locator('div[role="button"]').filter({ hasText: traceId }).first();

    // Retry logic for finding the trace, as the backend might lag
    await expect(async () => {
        // Only reload if not visible
        if (!await traceItem.isVisible()) {
             await page.reload();
             await searchInput.fill(traceId);
        }
        await expect(traceItem).toBeVisible({ timeout: 5000 });
    }).toPass({ timeout: 30000 });

    await traceItem.click();

    // Verify detail view
    // The details pane should show the Trace ID
    await expect(page.getByText(traceId)).toBeVisible();

    // Verify visualization components are present
    await expect(page.getByText('Execution Waterfall')).toBeVisible();
    await expect(page.getByText('Sequence Diagram')).toBeVisible();

    // Verify grouping: We expect a virtual root because we sent multiple roots
    await expect(page.getByText('Trace Group')).toBeVisible();

    // Check status
    await expect(page.getByText('SUCCESS').first()).toBeVisible();
  });
});
