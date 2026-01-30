/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Trace Sequence Diagram', () => {
  test('should display dynamic sequence diagram for a tool call', async ({ page, request }) => {
    // 1. Seed a trace by making a JSON-RPC call to the backend
    const backendUrl = process.env.BACKEND_URL || 'http://localhost:50050';
    console.log(`Using backend URL: ${backendUrl}`);

    // Send a tool call request
    const response = await request.post(`${backendUrl}/mcp`, {
      headers: {
        'Content-Type': 'application/json'
      },
      data: {
        jsonrpc: '2.0',
        method: 'tools/call',
        params: {
          name: 'get_weather',
          arguments: { location: 'E2E_Test_City' }
        },
        id: 12345
      }
    });

    expect(response.status()).toBe(200);

    // 2. Navigate to traces page
    await page.goto('/traces');

    // 3. Wait for the trace to appear and click it
    // We might need to refresh as the page loads traces on mount
    // Or wait a bit for the backend to flush to the ring buffer (it should be instant)
    await page.reload();

    // Look for the trace. Our API constructs the name as "Execute Request" for the root span of a tool call,
    // but the list item title comes from trace.rootSpan.name which we set to "Execute Request".
    // Wait, let's check TraceList component.
    // ui/src/components/traces/trace-list.tsx might use trace.rootSpan.name

    // Actually, in route.ts:
    // rootSpan.name = "Execute Request";
    // So we look for "Execute Request" in the list.

    // Wait, if multiple traces exist, we need the specific one.
    // But "Execute Request" is generic.
    // Maybe we can filter or just pick the first one which should be the latest (sorted by timestamp desc).

    await expect(page.getByText('Execute Request').first()).toBeVisible();
    await page.getByText('Execute Request').first().click();

    // 4. Check Sequence Diagram
    // Participants
    await expect(page.getByText('Client', { exact: true })).toBeVisible();
    await expect(page.getByText('MCP Core', { exact: true })).toBeVisible();
    await expect(page.getByText('Tool', { exact: true })).toBeVisible();

    // Interactions
    // 1. Client -> Core: "Execute Request"
    await expect(page.locator('text="Execute Request"').nth(1)).toBeVisible(); // nth(1) because list item also has it

    // 2. Core -> Tool: "get_weather"
    // Use first() to avoid ambiguity with Waterfall view, or scope to SVG
    await expect(page.locator('svg').getByText('get_weather', { exact: true })).toBeVisible();

    // 3. Tool -> Core: "get_weather Result"
    await expect(page.getByText('get_weather Result')).toBeVisible();

    // 4. Core -> Client: "Execute Request Result" (Wait, code says label: `${span.name} Result`)
    // If span.name is "Execute Request", then label is "Execute Request Result".
    await expect(page.getByText('Execute Request Result')).toBeVisible();

    // Take verification screenshot
    // await page.screenshot({ path: 'verification_sequence.png', fullPage: true });
  });
});
