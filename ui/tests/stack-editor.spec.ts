/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Stack Editor', () => {
  test('should load the editor and show initial config', async ({ page }) => {
    // Navigate to a stack page (mocking the ID)
    await page.goto('/stacks/default-stack');

    // Check if the editor container is present
    const editor = page.locator('.monaco-editor');
    await expect(editor).toBeVisible({ timeout: 15000 });

    const visualizer = page.locator('.stack-visualizer-container');
    await expect(visualizer.getByText('weather-service', { exact: true })).toBeVisible();
  });

  // Enabled previously flaky interaction test
  test('should update visualizer when template added', async ({ page }) => {
    await page.goto('/stacks/default-stack');
    await page.locator('.monaco-editor').waitFor({ state: 'visible' });
    // Click a template in the sidebar
    await page.getByText('Redis', { exact: true }).click({ force: true });
    // Wait for updates
    await page.waitForTimeout(2000);
    const visualizer = page.locator('.stack-visualizer-container');
    await expect(visualizer.getByText('redis-cache')).toBeVisible({ timeout: 10000 });
  });
});
