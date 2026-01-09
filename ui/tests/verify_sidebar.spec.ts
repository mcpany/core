/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test('verify sidebar navigation', async ({ page }) => {
  // Go to homepage
  await page.goto('http://localhost:9002/');

  // Check if sidebar is visible
  await expect(page.locator('button[data-sidebar="trigger"]')).toBeVisible();

  // Check if "Platform" group exists
  await expect(page.getByText('Platform')).toBeVisible();

  // Check links
  await expect(page.getByRole('link', { name: 'Dashboard' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Network Graph' })).toBeVisible();

  // Take screenshot
  await page.screenshot({ path: `.audit/ui/${new Date().toISOString().split('T')[0]}/unified_navigation_system.png` });
});
