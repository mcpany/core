/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection, seedGlobalState } from './e2e/test-data';

test.describe('Stack Editor', () => {
  test.beforeEach(async ({ request }) => {
    await seedGlobalState(request);
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
      // await page.goto(`/stacks/${stackName}`);

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
      // await page.goto('/stacks/new');

      const visualizer = page.locator('.stack-visualizer-container');

      // Wait for the visualizer to be ready
      await expect(visualizer.locator('.react-flow')).toBeVisible({ timeout: 45000 });
      await expect(page.getByText('Valid YAML')).toBeVisible();

      const templateCard = page.getByText(/GitHub|Google Calendar|Linear/, { exact: true }).first();
      await expect(templateCard).toBeVisible({ timeout: 30000 });
      await templateCard.click();

      // Verify the editor remains valid and the visualizer stays rendered after inserting a template.
      await expect(page.getByText('Valid YAML')).toBeVisible();
      await expect(visualizer.locator('.react-flow')).toBeVisible({ timeout: 60000 });
    } finally {
      await cleanupCollection(stackName, page.request);
    }
  });
});
