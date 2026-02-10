/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

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
  // Ensure directory exists
  const dateDir = new Date().toISOString().split('T')[0];
  const fs = await import('fs');
  const path = await import('path');
  // Use test-results directory which should be writable
  const dir = path.join('test-results', 'audit', 'ui', dateDir);
  if (!fs.existsSync(dir)){
      fs.mkdirSync(dir, { recursive: true });
  }
  await page.screenshot({ path: path.join(dir, 'unified_navigation_system.png') });
});
