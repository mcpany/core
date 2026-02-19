/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import { seedUser } from './e2e/test-data';

const USER_ID = 'audit-admin-e2e';

test.describe('Visual Audit', () => {
  test.beforeEach(async ({ page, request }) => {
    // 1. Seed user
    await seedUser(request, USER_ID);

    // 2. Login
    await page.goto('/login');
    await page.getByLabel('Username').fill(USER_ID);
    await page.getByLabel('Password').fill('password'); // From test-data hash
    await page.getByRole('button', { name: 'Sign in' }).click({ force: true });
    await page.waitForURL('/', { timeout: 30000 });
  });

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
        // Use test-results directory which is writable in CI
        const screenshotPath = path.resolve(process.cwd(), `test-results/artifacts/audit/ui/${pageInfo.name}.png`);
        // eslint-disable-next-line @typescript-eslint/no-require-imports
        const fs = require('fs');
        try {
            fs.mkdirSync(path.dirname(screenshotPath), { recursive: true });
            await page.screenshot({ path: screenshotPath, fullPage: true });
        } catch (e) {
            console.warn('Failed to save screenshot:', e);
        }
      }
    });
  }
});
