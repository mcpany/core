/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stack Editor', () => {
  test.beforeEach(async ({ request }) => {
      await seedCollection('default-stack', request);
      // Wait a bit for potential backend sync (though seedCollection awaits response)
  });

  test.afterEach(async ({ request }) => {
      await cleanupCollection('default-stack', request);
  });

  test('should load the editor and show initial config in graph', async ({ page }) => {
    await page.goto('/stacks/default-stack');

    // Check for React Flow container
    const visualizer = page.locator('.stack-visualizer-container');
    await expect(visualizer.locator('.react-flow')).toBeVisible({ timeout: 30000 });

    // Check for the node
    // Using a more specific selector to ensure it's inside a node
    // Relaxed: check for text first
    await expect(visualizer).toContainText('weather-service', { timeout: 30000 });
    const weatherNode = visualizer.locator('.react-flow__node').filter({ hasText: 'weather-service' }).first();
    await expect(weatherNode).toBeVisible();
  });

  test('should update graph when template added', async ({ page }) => {
    await page.goto('/stacks/default-stack');
    const visualizer = page.locator('.stack-visualizer-container');

    // Wait for initial load with increased timeout and handling for slow rendering
    // Sometimes the text might be inside a child element
    // Relaxed check: Look for the text anywhere in the visualizer first
    // Note: If seeding failed, this will still timeout, but it's the best we can do without real backend logs in test runner.
    await expect(visualizer).toContainText('weather-service', { timeout: 45000 });

    // If text is found but node selector fails, it might be structure issue.
    // We already checked text, so we can be slightly lenient on the specific class selector if needed,
    // but ReactFlow usually has this class.
    // Try to find ANY node if specific one fails? No, specific is better.
    await expect(visualizer.locator('.react-flow__node').filter({ hasText: 'weather-service' }).first()).toBeVisible({ timeout: 10000 });

    // Click on PostgreSQL template in the palette
    // Ensure palette is visible first
    await expect(page.getByText('PostgreSQL')).toBeVisible();
    await page.getByText('PostgreSQL').click();

    // Verify new node appears in graph
    const postgresNode = visualizer.locator('.react-flow__node').filter({ hasText: 'postgres-db' });
    await expect(postgresNode).toBeVisible({ timeout: 60000 });
  });
});
