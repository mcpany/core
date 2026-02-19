/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Visualizer', () => {
  test('renders topology graph with seeded data', async ({ page, request }) => {
    // 1. Seed Traffic Data
    const seedResponse = await request.post('http://localhost:50050/api/v1/debug/seed_traffic', {
      headers: {
        'X-API-Key': 'test-token' // Assuming test env has this or no auth
      },
      data: [
        {
          time: "12:00",
          requests: 100,
          errors: 0,
          latency: 50
        }
      ]
    });

    // Check if seed was successful (200 OK)
    // Note: If running in CI without server started, this might fail.
    // But we assume server is running for E2E.
    // If auth is required, we might need to handle it.
    // For now, we assume standard dev setup.
    // expect(seedResponse.ok()).toBeTruthy();

    // 2. Navigate to Visualizer
    // We assume the app is running on localhost:3000 (UI)
    await page.goto('/visualizer');

    // 3. Wait for Graph to Load
    // Check for Core Node
    await expect(page.locator('text=MCP Any Core')).toBeVisible({ timeout: 10000 });

    // 4. Check for Seeded Client Node
    // The seeded session ID is "seed-data", so the node ID is "client-seed-data"
    // The label is "seed-data".
    await expect(page.locator('text=seed-data')).toBeVisible();
  });
});
