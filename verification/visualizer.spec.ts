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
    // Note: We need to authenticate. If the test env has auth disabled or we have a key.
    // Assuming default dev environment or we can bypass if needed.
    // The previous analysis showed authMiddleware is active but allows localhost if no API key.
    // Playwright runs in the same container usually or network.

    // We can't easily fetch from Playwright context to API unless we use request context.
    const apiContext = await page.context().request;
    const seedRes = await apiContext.post(`${API_URL}/api/v1/debug/seed_traffic`, {
        data: trafficPoints,
        headers: {
            // "X-API-Key": "test-key" // If needed
        }
    });

    // We expect 200 or 401/403. If 401/403 we might need to handle auth.
    // But for "localhost" access without configured key, it should pass as admin.
    expect(seedRes.ok()).toBeTruthy();

    // 2. Navigate to Visualizer
    await page.goto('/visualizer');

    // 3. Verify Graph Renders
    // React Flow renders nodes with class 'react-flow__node'
    await expect(page.locator('.react-flow__node')).not.toHaveCount(0);

    // Check for "MCP Any" Core Node (label text)
    await expect(page.getByText('MCP Any')).toBeVisible();

    // Check for "Live" toggle
    await expect(page.getByText('Live')).toBeVisible();

    // 4. Verify Edges exist (implies graph connectivity)
    // await expect(page.locator('.react-flow__edge')).not.toHaveCount(0);
  });
});
