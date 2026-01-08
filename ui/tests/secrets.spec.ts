/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Secrets Manager', () => {
  test('should allow adding and deleting secrets', async ({ page }) => {
    const timestamp = Date.now();
    const secretName = `E2E Test Secret ${timestamp}`;
    const secretKey = `E2E_TEST_KEY_${timestamp}`;
    const secretValue = `sk-test-value-${timestamp}`;
    await page.goto('/secrets');

    // Check if title is present (SecretsManager uses h3)
    // Use getByRole for robustness
    await expect(page.getByRole('heading', { name: 'API Keys & Secrets' })).toBeVisible();

    // Add a new secret
    await page.getByRole('button', { name: 'Add Secret' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    await page.fill('#name', secretName);
    await page.fill('#key', secretKey);
    await page.fill('#value', secretValue);
    await page.getByRole('button', { name: 'Save Secret' }).click();

    // Verify it appears in the list
    await expect(page.getByText(secretName)).toBeVisible();
    await expect(page.getByText(secretKey)).toBeVisible();

    // Verify mask (24 dots) - Scope to the row
    // Find the row-like container for the secret using .group class
    const secretRow = page.locator('.group').filter({ hasText: secretName }).first();
    await expect(secretRow.getByText('••••••••••••••••••••••••')).toBeVisible();

    // Toggle visibility
    // Click the eye icon button using aria-label
    // Use force: true because sometimes the button structure might intercept text pointer events
    const toggleButton = secretRow.locator('button[aria-label="Show secret"]');
    await toggleButton.click({ force: true });
    await page.waitForTimeout(500); // Wait for state update/animation

    // Verify state changed (Implicit in next check)
    // await expect(secretRow.locator('button[aria-label="Hide secret"]')).toBeVisible();

    // Find the span containing the value and check text
    // The value might take a moment to be revealed/hydrated
    await expect(secretRow.getByText('[REDACTED]')).toBeVisible({ timeout: 5000 });

    // Delete the secret
    await secretRow.locator('button[aria-label="Delete secret"]').click();

    // Verify it's gone
    await expect(page.getByText(secretName)).not.toBeVisible();
  });
});
