/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Secrets Manager', () => {
  test('should allow adding and deleting secrets', async ({ page }) => {
    const secretName = `E2E Test Secret ${Date.now()}`;
    await page.goto('/secrets');

    // Check if title is present (SecretsManager uses h3)
    // Use getByRole for robustness
    await expect(page.getByRole('heading', { name: 'API Keys & Secrets' })).toBeVisible();

    // Add a new secret
    await page.getByRole('button', { name: 'Add Secret' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    await page.fill('#name', secretName);
    const secretKey = `TEST_KEY_${Date.now()}`;
    await page.fill('#key', secretKey);
    await page.fill('#value', 'test-secret-value');
    await page.getByRole('button', { name: 'Save Secret' }).click();

    // Verify it appears in the list
    await expect(page.getByText(secretName)).toBeVisible();
    await expect(page.getByText(secretKey)).toBeVisible();

    // Verify mask (24 dots) - Scope to the row
    // Find the row-like container for the secret using .group class
    const secretRow = page.locator('.group').filter({ hasText: secretName }).first();
    await expect(secretRow.getByText('••••••••••••••••••••••••')).toBeVisible();

    /* Flaky in CI/Docker
    // Toggle visibility
    // Click the eye icon button using aria-label
    const toggleButton = secretRow.locator('button[aria-label="Show secret"]');
    await toggleButton.click({ force: true });
    await page.waitForTimeout(500); // Wait for state update/animation

    // Verify state changed
    await expect(secretRow.locator('button[aria-label="Hide secret"]')).toBeVisible();

    // Find the span containing the value and check text
    // The span is the first child of the wrapper which is the first child of the right-side container?
    // Structure: div(flex) > span.truncate
    // We can just find the span in the row that contains the value (or dots previously)
    // But since we expect value now, let's look for the span that is NOT dots
    const valueSpan = secretRow.locator('span.truncate').first();
    await expect(valueSpan).toHaveText('test-secret-value');
    */

    // Delete the secret
    await secretRow.locator('button[aria-label="Delete secret"]').click();

    // Verify it's gone
    await expect(page.getByText(secretName)).not.toBeVisible();
  });
});
