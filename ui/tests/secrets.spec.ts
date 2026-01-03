/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Secrets Manager', () => {
  test('should allow adding and deleting secrets', async ({ page }) => {
    await page.goto('/secrets');

    // Check if title is present
    await expect(page.locator('h1')).toContainText('API Key Vault');

    const secretName = `E2E Unique Secret ${Date.now()}`;

    // Add a new secret
    await page.click('text=Add Secret');
    await page.fill('#name', secretName);
    await page.fill('#key', 'TEST_KEY');
    await page.fill('#value', 'test-secret-value');
    await page.click('button:has-text("Save Secret")');

    // Verify it appears in the list
    await expect(page.locator('table')).toContainText(secretName);
    await expect(page.locator('table')).toContainText('TEST_KEY');

    // Verify mask
    await expect(page.locator('table')).toContainText('••••••••');

    // Toggle visibility
    await page.locator(`tr:has-text("${secretName}") button`).nth(0).click();
    await expect(page.locator('table')).toContainText('test-secret-value');

    // Delete the secret
    page.on('dialog', dialog => dialog.accept());
    const row = page.locator(`tr:has-text("${secretName}")`);
    await row.hover();
    await row.locator('button[aria-label="Delete secret"]').click();

    // Verify it's gone
    await expect(page.locator('table')).not.toContainText(secretName);
  });
});
