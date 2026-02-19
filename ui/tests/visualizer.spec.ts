/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Agent Flow Visualizer', () => {
  const backendUrl = 'http://localhost:50050';

  test.beforeAll(async () => {
    // Seed traffic data to ensure the graph is populated
    const seedData = [
      {
        time: "10:00",
        requests: 100,
        errors: 0,
        latency: 50
      },
      // We are seeding traffic history which triggers session creation in the backend
      // for "seed-data" session ID.
    ];

    const response = await fetch(`${backendUrl}/api/v1/debug/seed_traffic`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': 'test-key' // Assuming test key or no auth in test env
      },
      body: JSON.stringify(seedData)
    });

    // If 401/403, we might need to handle auth, but for now assume dev/test env config
    if (response.status === 401 || response.status === 403) {
        console.warn("Seeding failed with auth error, proceeding but graph might be empty if not handled");
    }
  });

  test('should render the topology graph with seeded data', async ({ page }) => {
    // Go to visualizer page
    await page.goto('/visualizer');

    // Wait for the graph container
    await expect(page.locator('.react-flow')).toBeVisible();

    // Check for the seeded client node
    // The seeded session ID is "seed-data" (from server/pkg/topology/manager.go)
    // The node ID will be "client-seed-data" (mapped in manager.go)
    // React Flow renders nodes with data-id attribute matching node ID
    await expect(page.locator('[data-id="client-seed-data"]')).toBeVisible({ timeout: 10000 });

    // Check for the Core node
    await expect(page.locator('[data-id="mcp-core"]')).toBeVisible();

    // Check for edge between client and core
    // Edge ID: "e-client-seed-data-mcp-core"
    // React Flow edges are SVG paths, harder to select by data-id sometimes depending on version,
    // but newer versions support it or we can check for existence of connection line.
    // For now, checking nodes is sufficient proof of graph rendering.
  });

  test('should toggle live mode', async ({ page }) => {
    await page.goto('/visualizer');

    const toggle = page.locator('label[for="live-mode"]');
    await expect(toggle).toBeVisible();

    // Default is off
    await expect(page.locator('#live-mode')).not.toBeChecked();

    // Click to toggle
    await toggle.click();
    await expect(page.locator('#live-mode')).toBeChecked();
  });
});
