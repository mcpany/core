/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Secrets Manager', () => {
  test('should allow adding and deleting secrets', async ({ page }) => {
    await page.goto('/secrets');

    // Check if title is present (SecretsManager uses h3)
    // Use getByRole for robustness
    await expect(page.getByRole('heading', { name: 'API Keys & Secrets' })).toBeVisible();

    // Add a new secret
    await page.getByRole('button', { name: 'Add Secret' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    await page.fill('#name', 'E2E Test Secret');
    await page.fill('#key', 'TEST_KEY');
    await page.fill('#value', 'test-secret-value');
    await page.getByRole('button', { name: 'Save Secret' }).click();

    // Verify it appears in the list
    await expect(page.getByText('E2E Test Secret')).toBeVisible();
    await expect(page.getByText('TEST_KEY')).toBeVisible();

    // Verify mask (24 dots)
    await expect(page.getByText('••••••••••••••••••••••••')).toBeVisible();

    // Toggle visibility
    // Find the row-like container for the secret using .group class
    const secretRow = page.locator('.group').filter({ hasText: 'E2E Test Secret' }).first();

    // Click the eye icon button - assume it is the first button in the row
    await secretRow.locator('button').first().click();
    await expect(page.getByText('test-secret-value')).toBeVisible();

    // Delete the secret
    await secretRow.locator('button[aria-label="Delete secret"]').click();

    // Verify it's gone
    await expect(page.getByText('E2E Test Secret')).not.toBeVisible();
  });
});
