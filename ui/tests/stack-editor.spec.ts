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
    // React Flow might take a moment to mount and render its internal structure
    // We check for the renderer wrapper or just wait for the container to be stable
    await expect(visualizer).toBeVisible();
    await expect(visualizer.locator('.react-flow__renderer')).toBeVisible({ timeout: 30000 });

    // Check for the node
    // Using a more specific selector to ensure it's inside a node
    const weatherNode = visualizer.locator('.react-flow__node').filter({ hasText: 'weather-service' });
    await expect(weatherNode).toBeVisible();
  });

  test('should update graph when template added', async ({ page }) => {
    await page.goto('/stacks/default-stack');
    const visualizer = page.locator('.stack-visualizer-container');

    // Wait for initial load
    await expect(visualizer.locator('.react-flow__node').filter({ hasText: 'weather-service' })).toBeVisible({ timeout: 30000 });

    // Click on PostgreSQL template in the palette
    await page.getByText('PostgreSQL').click();

    // Verify new node appears in graph
    const postgresNode = visualizer.locator('.react-flow__node').filter({ hasText: 'postgres-db' });
    await expect(postgresNode).toBeVisible({ timeout: 10000 });
  });
});
