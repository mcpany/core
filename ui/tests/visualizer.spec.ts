/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Network Topology Visualizer', () => {
  test('should render and load seeded topology data', async ({ page }) => {
    // 1. Go to Visualizer page
    await page.goto('/visualizer');

    // 2. Click Seed Data button
    await page.getByRole('button', { name: 'Seed' }).click();

    // 3. Verify Toast appears
    await expect(page.getByText('Traffic Seeded', { exact: true })).toBeVisible();

    // 4. Verify Nodes appear
    // The Core node ID is "mcp-core"
    await expect(page.getByTestId('rf__node-mcp-core')).toBeVisible();

    // "seed-data" is the session ID created by seed function
    // Client node ID format: "client-{id}" -> "client-seed-data"
    await expect(page.getByTestId('rf__node-client-seed-data')).toBeVisible();

    // 5. Verify Edges
    const edges = page.locator('.react-flow__edge');
    // We expect at least one edge (Client -> Core).
    // There might be more due to default middlewares and services.
    await expect(edges).toHaveCount(await edges.count());
    const count = await edges.count();
    expect(count).toBeGreaterThan(0);

    // 6. Screenshot
    await page.screenshot({ path: '../verification/visualizer.png', fullPage: true });
  });
});
