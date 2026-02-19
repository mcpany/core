/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Visualizer', () => {
  test.beforeEach(async ({ request }) => {
    // Seed dummy traffic data to ensure the graph is populated
    const points = [
      {
        time: "10:00",
        requests: 100,
        errors: 0,
        latency: 50
      }
    ];
    // This call updates the "seed-data" session in the backend
    const res = await request.post('/api/v1/debug/seed_traffic', {
      data: points
    });
    expect(res.ok()).toBeTruthy();
  });

  test('should display topology graph with core and client nodes', async ({ page }) => {
    await page.goto('/visualizer');

    // Wait for canvas to load
    await page.waitForSelector('.react-flow__renderer');

    // Check for "MCP Gateway" node (Core)
    // React Flow nodes usually have text content
    await expect(page.getByText('MCP Gateway')).toBeVisible();

    // Check for "seed-data" client node (created by seeding)
    // The visualizer might show "seed-data" or "Client" depending on label logic.
    // In manager.go: label = session.ID ("seed-data").
    // In agent-flow.tsx: data: { label: client.label || 'Client' }
    // So it should be "seed-data".
    await expect(page.getByText('seed-data')).toBeVisible();
  });
});
