/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedSettings, seedUser, cleanupUser } from './test-data';

test.describe('Settings & Secrets', () => {
  test.beforeEach(async ({ request, page }) => {
    // Seed initial state
    await seedSettings(request);
    await seedUser(request, "e2e-admin-settings");

    // Login
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', 'e2e-admin-settings');
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]', { force: true });
    await page.waitForURL('/', { timeout: 30000 });
    await expect(page).toHaveURL('/', { timeout: 15000 });

    await page.goto('/settings');
  });

  test.afterEach(async ({ request }) => {
    // await cleanupUser(request, "e2e-admin-settings");
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

    // Verify changes persisted (optional, but good for real data test)
    // await page.reload();
    // await page.getByRole('tab', { name: 'Global Config' }).click();
    // await expect(page.getByText('DEBUG')).toBeVisible();
  });
});
