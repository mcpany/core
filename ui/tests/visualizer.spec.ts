/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect, request } from '@playwright/test';

test.describe('Agent Flow Visualizer', () => {
  test.beforeAll(async ({ playwright }) => {
    // Seed traffic data to ensure the graph is populated.
    // We use the Playwright request context to leverage the baseURL from config (Frontend URL).
    // This hits the Next.js API proxy which forwards to the backend, avoiding direct backend connection issues.
    const apiContext = await playwright.request.newContext();

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

    const response = await apiContext.post('/api/v1/debug/seed_traffic', {
      headers: {
        'Content-Type': 'application/json',
        // X-API-Key is already injected by playwright.config.ts extraHTTPHeaders
      },
      data: seedData
    });

    if (!response.ok()) {
      console.warn(`Seeding failed with status ${response.status()}: ${await response.text()}`);
    }

    await apiContext.dispose();
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
