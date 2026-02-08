/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';

test.describe('Visual Audit', () => {
  const pages = [
    { name: 'dashboard', path: '/' },
    { name: 'stacks_list', path: '/stacks' },
    { name: 'stack_detail', path: '/stacks/system' }, // Dummy ID
    { name: 'services_list', path: '/upstream-services' },
    { name: 'settings', path: '/settings' },
  ];

  for (const pageInfo of pages) {
    test(`capture ${pageInfo.name}`, async ({ page }) => {
      await page.goto(pageInfo.path);
      // Wait for some basic element to ensure load, though layout should be enough
      await expect(page.locator('body')).toBeVisible();

      // Additional wait for "hydration" or animations if needed
      await page.waitForTimeout(1000);

      if (process.env.CAPTURE_SCREENSHOTS === 'true') {
        const screenshotPath = path.resolve(__dirname, `../../.audit/ui/${pageInfo.name}.png`);
        await page.screenshot({ path: screenshotPath, fullPage: true });
      }
    });
  }
});
