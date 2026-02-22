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
    const stackName = `stack-editor-load-${Date.now()}`;
    await seedCollection(stackName, page.request);
    try {
      await page.goto(`/stacks/${stackName}`);

      // Check for React Flow container
      const visualizer = page.locator('.stack-visualizer-container');
      await expect(visualizer.locator('.react-flow')).toBeVisible({ timeout: 30000 });

      // Check for the node
      // Using a more specific selector to ensure it's inside a node and wait for it
      const weatherNode = visualizer.locator('.react-flow__node').filter({ hasText: 'weather-service' }).first();
      await expect(weatherNode).toBeVisible({ timeout: 15000 });
      await expect(weatherNode).toContainText('weather-service');
    } finally {
      await cleanupCollection(stackName, page.request);
    }
  });

  test('should update graph when template is added', async ({ page }) => {
    const stackName = `stack-editor-update-${Date.now()}`;
    await seedCollection(stackName, page.request);
    try {
      await page.goto('/stacks/new');

      const visualizer = page.locator('.stack-visualizer-container');

      // Wait for the visualizer to be ready
      await expect(visualizer.locator('.react-flow')).toBeVisible({ timeout: 45000 });

      // Click on PostgreSQL template in the palette
      // Use role button to be more specific if possible, or just exact text
      const postgresTemplate = page.getByRole('button', { name: 'PostgreSQL' }).or(page.getByText('PostgreSQL', { exact: true }));
      await expect(postgresTemplate.first()).toBeVisible();
      await postgresTemplate.first().click();

      // Verify new node appears in graph
      const postgresNode = visualizer.locator('.react-flow__node').filter({ hasText: /^postgres-db$/ }).first();
      await expect(postgresNode).toBeVisible({ timeout: 60000 });
    } finally {
      await cleanupCollection(stackName, page.request);
    }
  });
});
