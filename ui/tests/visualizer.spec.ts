/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Visualizer', () => {
  test('should display live topology graph with seeded data', async ({ page, request }) => {
    // 1. Seed Traffic Data
    // This creates a dummy session "seed-data" in the backend topology manager
    const seedResponse = await request.post('/api/v1/debug/seed_traffic', {
      data: [
        {
          time: '12:00',
          requests: 100,
          errors: 0,
          latency: 50
        }
      ]
    });
    expect(seedResponse.ok()).toBeTruthy();

    // 2. Navigate to Visualizer
    await page.goto('/visualizer');

    // 3. Wait for Graph to load
    // The "Loading..." text should disappear
    await expect(page.getByText('Loading topology...')).not.toBeVisible({ timeout: 15000 });

    // 4. Assert Nodes using React Flow test IDs or specific text scoping
    // The ID for the core node is "mcp-core", React Flow prefixes it with "rf__node-"
    await expect(page.getByTestId('rf__node-mcp-core')).toBeVisible();
    await expect(page.getByTestId('rf__node-mcp-core')).toContainText('MCP Any');
    await expect(page.getByTestId('rf__node-mcp-core')).toContainText('Gateway');

    // The seed data creates a session named "seed-data"
    // Client node ID format is "client-" + sessionID -> "client-seed-data"
    await expect(page.getByTestId('rf__node-client-seed-data')).toBeVisible();
    await expect(page.getByTestId('rf__node-client-seed-data')).toContainText('seed-data');
  });
});
