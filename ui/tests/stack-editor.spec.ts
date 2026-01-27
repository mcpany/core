/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stack Editor', () => {
  test.beforeEach(async ({ request }) => {
      await seedCollection('default-stack', request);
      const res = await request.get('/api/v1/collections/default-stack', { headers: { 'X-API-Key': 'test-token' } });
      console.log(`VERIFY COLLECTION: ${res.status()} ${await res.text()}`);
  });

  test.afterEach(async ({ request }) => {
      await cleanupCollection('default-stack', request);
  });

  test('should load the editor and show initial config', async ({ page }) => {
    await page.goto('/stacks/default-stack');
    const visualizer = page.locator('.stack-visualizer-container');
    await expect(visualizer.getByText('weather-service', { exact: true })).toBeVisible({ timeout: 15000 });
  });

  test('should update visualizer when template added', async ({ page }) => {
    await page.goto('/stacks/default-stack');
    const visualizer = page.locator('.stack-visualizer-container');
    await expect(visualizer.getByText('weather-service', { exact: true })).toBeVisible({ timeout: 15000 });

    await page.getByText('Redis').click();
    await expect(visualizer.getByText('redis-cache', { exact: true })).toBeVisible();
  });
});
