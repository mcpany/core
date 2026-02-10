/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';

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
  const date = new Date().toISOString().split('T')[0];
  // Use test-results to ensure write permissions in CI
  const auditDir = path.join('test-results', '.audit/ui', date);
  if (!fs.existsSync(auditDir)) {
      try {
        fs.mkdirSync(auditDir, { recursive: true });
      } catch (e) {
        console.warn(`Could not create audit directory: ${e}. Skipping screenshot save.`);
        return;
      }
  }
  try {
    await page.screenshot({ path: path.join(auditDir, 'unified_navigation_system.png') });
  } catch (e) {
    console.warn(`Failed to take screenshot: ${e}`);
  }
});
