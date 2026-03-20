/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

// Mock trace matching the shape that generateMockTrace() produces on the backend.
const MOCK_TRACE = {
  id: 'trace-seed-inspector-test',
  timestamp: new Date().toISOString(),
  totalDuration: 1250,
  status: 'success',
  trigger: 'user',
  rootSpan: {
    id: 'span-orchestrator-1',
    name: 'orchestrator-task',
    type: 'service',
    status: 'success',
    duration: 1250,
    children: [
      {
        id: 'span-child-1',
        name: 'fetch-data',
        type: 'tool',
        status: 'success',
        duration: 300,
        children: [],
      },
    ],
  },
};

test.describe('Inspector Page', () => {
  test('should allow seeding a trace from backend and viewing it', async ({ page }) => {
    // Intercept the POST request to /api/v1/debug/traces and simulate backend response
    await page.route('**/api/v1/debug/traces', async (route) => {
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({ status: 'seeded', id: MOCK_TRACE.id }),
      });
    });

    // Intercept the WebSocket connection for traces.
    // vite preview does not forward WebSocket upgrades through its proxy, so we
    // mock the WS at the browser level to ensure the trace is delivered to the
    // InspectorTable without depending on proxy-level WS tunnelling.
    let wsSend: ((data: string) => void) | null = null;
    await page.routeWebSocket('**/api/v1/ws/traces', (ws: any) => {
      wsSend = (data: string) => ws.send(data);
    });

    // Navigate to the Inspector page
    await page.goto('/inspector');

    // Wait for the page to load by checking for the "Inspector" header
    await expect(page.getByRole('heading', { name: 'Inspector' })).toBeVisible();

    // Click the "Seed Trace" button (triggers POST /api/v1/debug/traces on backend)
    const seedTraceBtn = page.getByRole('button', { name: 'Seed Trace' });
    await expect(seedTraceBtn).toBeVisible();
    await seedTraceBtn.click();

    // Expect the toast notification confirming the backend received the seed request
    await expect(page.getByText('Trace Seeded').first()).toBeVisible({ timeout: 5000 });

    // After the POST succeeds, inject the trace into the active WebSocket
    // connection.
    if (wsSend && typeof wsSend === 'function') {
      (wsSend as any)(JSON.stringify(MOCK_TRACE));
    }

    // Wait briefly to allow React state to update based on WebSocket message
    await page.waitForTimeout(500);

    // The injected trace's root span name should appear in the inspector table.
    const row = page.locator('text=orchestrator-task').first();
    await expect(row).toBeVisible({ timeout: 10000 });

    // Click the row to open the detail sheet
    await row.click();

    // Verify the detail sheet opens and shows trace info
    const sheet = page.getByRole('dialog');
    await expect(sheet).toBeVisible();
    await expect(sheet.locator('text=orchestrator-task').first()).toBeVisible();
  });
});
