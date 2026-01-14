/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Middleware Page', () => {
  test('should open configuration sheet', async ({ page }) => {
    await page.goto('/middleware');

    // Find the row containing "Rate Limiter" in the Active Pipeline list
    // The row has class "flex items-center justify-between" and contains "Rate Limiter"
    const row = page.locator('.flex.items-center.justify-between').filter({ hasText: 'Rate Limiter' }).first();
    const settingsBtn = row.getByRole('button').last(); // Settings button is the last one

    // Ensure row is visible
    await expect(row).toBeVisible();

    // Click settings
    await settingsBtn.click();

    // Check for Sheet Content
    // "Configure Middleware" should be visible in the dialog
    const sheet = page.getByRole('dialog');
    await expect(sheet).toBeVisible();
    await expect(sheet.getByText('Configure Middleware')).toBeVisible();
    await expect(sheet.getByText('Rate Limiter', { exact: false })).toBeVisible();
  });
});
