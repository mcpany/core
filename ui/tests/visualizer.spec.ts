/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Network Topology Visualizer', () => {
  test.beforeEach(async ({ page, request }) => {
    // 1. Seed Traffic Data via API
    // We use the request fixture to call the backend API directly
    // time format expected by backend is "15:04"
    const now = new Date();
    const time1 = new Date(now.getTime() - 2 * 60000);
    const timeStr1 = `${String(time1.getHours()).padStart(2, '0')}:${String(time1.getMinutes()).padStart(2, '0')}`;
    const time2 = new Date(now.getTime() - 1 * 60000);
    const timeStr2 = `${String(time2.getHours()).padStart(2, '0')}:${String(time2.getMinutes()).padStart(2, '0')}`;

    const trafficData = [
      { time: timeStr1, total: 100, errors: 2, latency: 50 },
      { time: timeStr2, total: 120, errors: 0, latency: 45 },
    ];

    // The backend URL is proxied via Next.js in dev, but in test we can hit the API route on Next.js
    // Playwright baseURL is configured to the Next.js app.
    const response = await request.post('/api/v1/debug/seed_traffic', {
      data: trafficData,
      headers: {
        'Authorization': 'Bearer test-token',
        'X-API-Key': 'test-token'
      }
    });

    if (!response.ok()) {
      console.log('Seed traffic failed:', await response.text());
    }
    expect(response.ok()).toBeTruthy();
  });

  test('should render topology graph with seeded data', async ({ page }) => {
    // 2. Navigate to Visualizer
    await page.goto('/visualizer');

    // 3. Verify Page Title
    await expect(page.getByRole('heading', { name: 'Agent Flow Visualizer' })).toBeVisible();

    // 4. Verify React Flow Canvas exists
    await expect(page.locator('.react-flow')).toBeVisible();

    // 5. Verify Core Node (MCP Any Gateway)
    // The label is "MCP Any" (hardcoded in manager.go)
    // React Flow nodes text content usually found in the node divs
    // We restrict search to the react-flow container to avoid matching sidebar title
    await expect(page.locator('.react-flow').getByText('MCP Any')).toBeVisible();

    // 6. Wait for refresh and verify graph update
    // Since seed traffic just updates history, it might not add new NODES unless we also register services.
    // The Core node is always there.
    // If we want to see Client nodes, the SeedTrafficHistory in backend updates SessionStats?
    // Let's check manager.go again.

    // Manager.go:
    // m.sessions["seed-data"] = ...
    // And GetGraph iterates sessions.
    // So "seed-data" session should appear as a Client Node!
    // The label for session "seed-data" will be "seed-data" (ID) unless metadata has userAgent.

    // So we expect to see "seed-data" or "Client" node.
    // My hook maps client.label or 'Client'.

    // Let's look for "seed-data" text on the canvas.
    // Wait for it to appear (React Flow might animate)
    await expect(page.getByText('seed-data')).toBeVisible({ timeout: 10000 });
  });
});
