/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

test('verify sidebar navigation', async ({ page }) => {
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
  const dateStr = new Date().toISOString().split('T')[0];
  // Align with audit-logs.spec.ts which uses ../.audit/ui relative to tests dir
  const screenshotDir = path.join(__dirname, '../.audit/ui', dateStr);

  if (!fs.existsSync(screenshotDir)) {
    fs.mkdirSync(screenshotDir, { recursive: true });
  }
  await page.screenshot({ path: path.join(screenshotDir, 'unified_navigation_system.png') });
});
