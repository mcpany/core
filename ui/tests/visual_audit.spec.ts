/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import { login } from './e2e/auth-helper';
import { seedUser, cleanupUser } from './e2e/test-data';

test.describe('Visual Audit', () => {
  test.beforeEach(async ({ page, request }) => {
    await seedUser(request, "e2e-admin");
    await login(page);
  });

  test.afterEach(async ({ request }) => {
    await cleanupUser(request, "e2e-admin");
  });

  const pages = [
    { name: 'dashboard', path: '/' },
    { name: 'stacks_list', path: '/stacks' },
    { name: 'stack_detail', path: '/stacks/system' },
    { name: 'services_list', path: '/upstream-services' },
    { name: 'settings', path: '/settings' },
  ];

  for (const pageInfo of pages) {
    test(`capture ${pageInfo.name}`, async ({ page }) => {
      await page.goto(pageInfo.path);
      // Wait for some basic element to ensure load
      await expect(page.locator('body')).toBeVisible();

      // Additional wait for "hydration" or animations if needed
      // Reduced from 1000 to 500 to save time in CI, assuming networkidle handles most
      await page.waitForTimeout(500);

      if (process.env.CAPTURE_SCREENSHOTS === 'true') {
        const screenshotPath = path.resolve(__dirname, `../../.audits/ui/${pageInfo.name}.png`);
        await page.screenshot({ path: screenshotPath, fullPage: true });
      }
    });
  }
});
