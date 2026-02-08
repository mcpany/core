/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import fs from 'fs';
import { login } from './e2e/auth-helper';
import { seedUser, cleanupUser, seedTraffic } from './e2e/test-data';

test.beforeEach(async ({ page, request }) => {
    await seedUser(request, "e2e-admin");
    await seedTraffic(request);
    await login(page);
});

test.afterEach(async ({ request }) => {
    await cleanupUser(request, "e2e-admin");
});

test('verify stats page', async ({ page }) => {
  // Go to the stats page
  await page.goto('/stats');

  // Wait for the dashboard to load
  await expect(page.getByText('Analytics & Stats')).toBeVisible();

  // Check for key elements
  await expect(page.getByText('Total Requests')).toBeVisible();
  await expect(page.getByText('Avg Latency')).toBeVisible();
  await expect(page.getByText('Error Rate')).toBeVisible();

  // Check tabs
  await expect(page.getByText('Overview')).toBeVisible();
  await expect(page.getByText('Performance')).toBeVisible();
  await expect(page.getByText('Errors')).toBeVisible();

  // Wait a bit for charts to animate (if any)
  await page.waitForTimeout(2000);

  // Take a screenshot
  if (process.env.CAPTURE_SCREENSHOTS === 'true') {
      const date = new Date().toISOString().split('T')[0];
      const screenshotDir = path.resolve(__dirname, '../.audit/ui', date);
      if (!fs.existsSync(screenshotDir)) {
          fs.mkdirSync(screenshotDir, { recursive: true });
      }
      await page.screenshot({ path: path.join(screenshotDir, 'stats_analytics.png'), fullPage: true });
  }
});
