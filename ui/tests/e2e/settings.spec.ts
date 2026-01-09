/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Settings & Secrets', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/settings');
  });

  test('should manage global settings and secrets', async ({ page }) => {
    // Global Settings (Log Level)
    await page.getByRole('tab', { name: 'General' }).click();
    const logLevelTrigger = page.getByRole('combobox').first();
    await expect(logLevelTrigger).toBeVisible();
    await logLevelTrigger.click();
    await page.getByRole('option', { name: 'DEBUG' }).click();
    await page.getByRole('button', { name: 'Save Settings' }).click();

    // Secrets Management
    await page.getByRole('tab', { name: 'Secrets & Keys' }).click();
    await page.getByRole('button', { name: 'Add Secret' }).click();

    const secretName = `test-secret-${Date.now()}`;
    await page.fill('input[id="name"]', secretName);
    await page.fill('input[id="key"]', 'TEST_KEY');
    await page.fill('input[id="value"]', 'TEST_VAL');

    await page.getByRole('button', { name: 'Save Secret' }).click();
    await expect(page.getByText(secretName)).toBeVisible();

    // Verify deletion
    const secretRow = page.locator('.group').filter({ hasText: secretName });
    page.on('dialog', dialog => dialog.accept());
    await secretRow.getByLabel('Delete secret').click();
    await expect(page.getByText(secretName)).not.toBeVisible();
  });
});
