/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import path from 'path';

test('verify sidebar navigation', async ({ page }, testInfo) => {
  // Go to homepage
  await page.goto('/');

  // Check if sidebar is visible
  await expect(page.locator('button[data-sidebar="trigger"]')).toBeVisible();

  // Check if "Platform" group exists
  await expect(page.getByText('Platform')).toBeVisible();

  // Check links
  await expect(page.getByRole('link', { name: 'Dashboard' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Network Graph' })).toBeVisible();

  // Take screenshot
  // Use testInfo.outputDir which is guaranteed to be writable and managed by Playwright
  const screenshotPath = path.join(testInfo.outputDir, 'unified_navigation_system.png');
  await page.screenshot({ path: screenshotPath });
});
