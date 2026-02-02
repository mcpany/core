import { test, expect } from '@playwright/test';
import crypto from 'crypto';

const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:50050';

test.describe('Distributed Tracing', () => {
  test('should visualize nested traces from W3C headers', async ({ page, request }) => {
    // Generate W3C compatible IDs
    // TraceID: 32 hex characters (16 bytes)
    // SpanID: 16 hex characters (8 bytes)
    const traceId = crypto.randomBytes(16).toString('hex');
    const parentSpanId = crypto.randomBytes(8).toString('hex');

    console.log(`Seeding trace: ${traceId}`);

    // Send two requests with the same Trace ID
    // This simulates a distributed trace where multiple spans belong to the same trace.
    // Both point to the same parent (simulating siblings).
    const headers = {
        'traceparent': `00-${traceId}-${parentSpanId}-01`,
        // Add fake user agent to be easily identifiable if needed, though traceId is enough
        'User-Agent': 'playwright-tracing-test'
    };

    await request.get(`${BACKEND_URL}/health`, { headers });
    await request.get(`${BACKEND_URL}/health`, { headers });

    // Wait for async processing in backend (ring buffer)
    // Increased wait time to ensure backend processes the log entry
    await new Promise(r => setTimeout(r, 2000));

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
        await page.reload(); // Reload to refresh list if live update is off/slow
        await searchInput.fill(traceId);
        await expect(traceItem).toBeVisible({ timeout: 5000 });
    }).toPass({ timeout: 20000 });

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
