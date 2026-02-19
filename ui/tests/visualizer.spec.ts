/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Topology Visualizer', () => {
  test('should display live topology graph with seeded data', async ({ page, request }) => {
    // 1. Seed Traffic Data to ensure we have active sessions
    const seedData = [
      {
        time: "12:00",
        requests: 100,
        errors: 0,
        latency: 50
      }
    ];

    const seedRes = await request.post('/api/v1/debug/seed_traffic', {
      data: seedData
    });
    expect(seedRes.ok()).toBeTruthy();

    // 2. Navigate to Visualizer
    await page.goto('/visualizer');

    // 3. Verify Nodes
    // We expect "MCP Any" (Core) inside a React Flow Node
    // Note: The text might be inside a child div, so we look for the node container
    await expect(page.locator('.react-flow__node').filter({ hasText: 'MCP Any' }).first()).toBeVisible();

    // We expect the seeded client (client-seed-data)
    await expect(page.locator('.react-flow__node').filter({ hasText: 'seed-data' }).first()).toBeVisible();

    // We expect "Gateway" role label
    await expect(page.locator('.react-flow__node').filter({ hasText: 'Gateway' }).first()).toBeVisible();
  });
});
