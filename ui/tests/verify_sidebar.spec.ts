/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

test('verify sidebar navigation', async ({ page }) => {
  // Navigate to home
  await page.goto('/');
  await expect(page).toHaveTitle(/MCPAny/);

  // Check sidebar elements
  const sidebar = page.locator('aside');
  await expect(sidebar).toBeVisible({ timeout: 30000 }); // Increased timeout

  // Verify links
  await expect(sidebar.getByRole('link', { name: 'Dashboard' })).toBeVisible();
  await expect(sidebar.getByRole('link', { name: 'Services' })).toBeVisible();
  await expect(sidebar.getByRole('link', { name: 'Tools' })).toBeVisible();
  await expect(sidebar.getByRole('link', { name: 'Resources' })).toBeVisible();
  await expect(sidebar.getByRole('link', { name: 'Prompts' })).toBeVisible();
  await expect(sidebar.getByRole('link', { name: 'Settings' })).toBeVisible();

  // Take screenshot
  const date = new Date().toISOString().split('T')[0];
  const screenshotDir = `.audit/ui/${date}`;
  if (!fs.existsSync(screenshotDir)) {
    fs.mkdirSync(screenshotDir, { recursive: true });
  }
  await page.screenshot({ path: `${screenshotDir}/unified_navigation_system.png` });
});
