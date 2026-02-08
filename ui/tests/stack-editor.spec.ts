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
    // Retry finding the node if it's not immediately available (React Flow can take a moment to render nodes)
    const weatherNode = visualizer.locator('.react-flow__node').filter({ hasText: 'weather-service' });
    // Increase timeout for node rendering
    await expect(weatherNode).toBeVisible({ timeout: 15000 });
  });

  test('should update graph when template added', async ({ page }) => {
    await page.goto('/stacks/default-stack');
    const visualizer = page.locator('.stack-visualizer-container');

    // Wait for initial load
    await expect(visualizer.locator('.react-flow__node').filter({ hasText: 'weather-service' })).toBeVisible({ timeout: 30000 });

    // Click on PostgreSQL template in the palette
    // Ensure palette is open or visible (default)
    const postgresTemplate = page.getByText('PostgreSQL');
    await expect(postgresTemplate).toBeVisible();
    await postgresTemplate.click();

    // Verify new node appears in graph
    const postgresNode = visualizer.locator('.react-flow__node').filter({ hasText: 'postgres-db' });
    // This might take time to re-layout
    await expect(postgresNode).toBeVisible({ timeout: 20000 });
  });
});
