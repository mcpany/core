/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from './test-data';

test.describe('Settings & Secrets', () => {
  const username = 'settings-admin';

  test.beforeEach(async ({ page, request }) => {
    // Seed a user for testing
    await seedUser(request, username);

    // Login
    await page.goto('/login');
    await page.fill('input[name="username"]', username);
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/');

    await page.goto('/settings');
  });

  test.afterEach(async ({ request }) => {
    await cleanupUser(request, username);
  });

  test('should manage global settings and secrets', async ({ page }) => {
    // 1. Global Settings (Log Level & Rate Limit)
    await page.getByRole('tab', { name: 'General' }).click();

    // Change Log Level to DEBUG
    const logLevelTrigger = page.locator('button[role="combobox"]').filter({ hasText: /INFO|WARN|ERROR|DEBUG/ }).first();
    await expect(logLevelTrigger).toBeVisible();
    // Only click if not already DEBUG (though seed usually defaults to INFO)
    await logLevelTrigger.click();
    await page.getByRole('option', { name: 'DEBUG' }).click();

    // Enable Rate Limit
    // The switch for "Global Rate Limit"
    const rateLimitSwitch = page.locator('button[role="switch"]').filter({ hasText: /Global Rate Limit/ }).first();
    // Assuming label is nearby or aria-label matches.
    // If we use shadcn Switch inside FormItem, the label is usually separate.
    // Let's target by text "Global Rate Limit" parent
    await page.getByLabel('Global Rate Limit').click();

    await page.getByRole('button', { name: 'Save Settings' }).click();
    // Wait for save to complete (button disabled then enabled)
    await expect(page.getByRole('button', { name: 'Saving...' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Save Settings' })).toBeEnabled();

    // Verify persistence by reloading
    await page.reload();
    await page.getByRole('tab', { name: 'General' }).click();
    await expect(page.getByRole('combobox').filter({ hasText: 'DEBUG' })).toBeVisible();
    await expect(page.getByLabel('Global Rate Limit')).toBeChecked();


    // 2. Secrets Management
    await page.getByRole('tab', { name: 'Secrets & Keys' }).click();

    await page.getByRole('button', { name: 'Add Secret' }).click();

    const secretName = `test-secret-${Date.now()}`;
    await page.fill('input[id="name"]', secretName);
    await page.fill('input[id="key"]', 'TEST_KEY');
    await page.fill('input[id="value"]', 'TEST_VAL');

    await page.getByRole('button', { name: 'Save Secret' }).click();
    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10000 });

    // Verify visibility
    await expect(page.getByText(secretName)).toBeVisible({ timeout: 10000 });

    // Verify deletion
    const secretRow = page.locator('.group').filter({ hasText: secretName });

    // Handle confirm dialog
    page.once('dialog', dialog => dialog.accept());

    await secretRow.getByLabel('Delete secret').click();

    await expect(page.getByText(secretName)).not.toBeVisible({ timeout: 10000 });
  });
});
