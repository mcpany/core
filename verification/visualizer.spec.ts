/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

// Use env var for BASE_URL or default to localhost:8080 (backend) and 3000 (frontend)
// E2E tests run against the Next.js app which proxies API requests to the backend.
const API_URL = process.env.API_URL || 'http://localhost:3000';

test.describe('Visualizer Topology', () => {
  test('should display real-time topology graph', async ({ page }) => {
    // 1. Seed Traffic Data
    const trafficPoints = [
      { time: "12:00", requests: 100, errors: 0, latency: 50 },
      { time: "12:01", requests: 150, errors: 2, latency: 60 }
    ];

    // Seed traffic via API
    const apiContext = await page.context().request;
    const seedRes = await apiContext.post(`${API_URL}/api/v1/debug/seed_traffic`, {
        data: trafficPoints,
        headers: {
            // "X-API-Key": "test-key" // If needed
        }
    });

    expect(seedRes.ok()).toBeTruthy();

    // 2. Navigate to Visualizer
    await page.goto('/visualizer');

    // 3. Verify Graph Renders
    // Check for "MCP Any" Core Node within the graph
    // React Flow assigns data-testid="rf__node-{id}"
    await expect(page.locator('[data-testid="rf__node-mcp-core"]')).toBeVisible();
    await expect(page.locator('[data-testid="rf__node-mcp-core"]').getByText('MCP Any')).toBeVisible();

    // Check for "Live" toggle
    await expect(page.getByLabel('Live')).toBeVisible();
  });
});
