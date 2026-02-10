/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stack Editor', () => {
  test('should load the editor and show initial config in graph', async ({ page, request }) => {
    const stackName = 'stack-editor-test-1';
    await seedCollection(stackName, request);
    try {
        await page.goto(`/stacks/${stackName}`);

        // Check for React Flow container
        const visualizer = page.locator('.stack-visualizer-container');
        await expect(visualizer.locator('.react-flow')).toBeVisible({ timeout: 30000 });

        // Check for the node
        // Using a more specific selector to ensure it's inside a node
        const weatherNode = visualizer.locator('.react-flow__node').filter({ hasText: 'weather-service' });
        await expect(weatherNode).toBeVisible();
    } finally {
        await cleanupCollection(stackName, request);
    }
  });

  // Skipping this test as it is covered by ui/tests/e2e/stack-composer.spec.ts which is more robust
  test.skip('should update graph when template added', async ({ page, request }) => {
    const stackName = 'stack-editor-test-2';
    await seedCollection(stackName, request);
    try {
        await page.goto(`/stacks/${stackName}`);
        const visualizer = page.locator('.stack-visualizer-container');

    // Wait for initial load with increased timeout and handling for slow rendering
    // Sometimes the text might be inside a child element
    await expect(visualizer.locator('.react-flow__node').filter({ hasText: 'weather-service' })).toBeVisible({ timeout: 45000 });

    // Click on PostgreSQL template in the palette
    // Ensure palette is visible first
    await expect(page.getByText('PostgreSQL')).toBeVisible();
    await page.getByText('PostgreSQL').click();

    // Verify new node appears in graph
    const postgresNode = visualizer.locator('.react-flow__node').filter({ hasText: 'postgres-db' });
    await expect(postgresNode).toBeVisible({ timeout: 60000 });
    } finally {
        await cleanupCollection(stackName, request);
    }
  });
});
