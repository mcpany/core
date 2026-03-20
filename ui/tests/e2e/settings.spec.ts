/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Settings & Secrets', () => {
  test.beforeEach(async ({ page, request }) => {
    await seedGlobalState(request);

    // Ensure login
    await page.goto('/login');
    await page.fill('input[name="username"]', 'e2e-admin-core'); // Matches seedGlobalState user
    await page.fill('input[name="password"]', 'password');
    await Promise.all([
      page.waitForURL('/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
    ]);

    await page.goto('/settings');
  });

  test('should manage global settings', async ({ page }) => {
    // Global Settings (Log Level)
    // "General" was renamed to "Global Config"
    await page.getByRole('tab', { name: 'Global Config' }).click();
    const logLevelTrigger = page.getByRole('combobox').first();
    await expect(logLevelTrigger).toBeVisible();
    await logLevelTrigger.click();
    await page.getByRole('option', { name: 'DEBUG' }).click();
    await page.getByRole('button', { name: 'Save Settings' }).click();
  });
});
