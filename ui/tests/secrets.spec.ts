/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Secrets Manager', () => {
  test('should allow adding and deleting secrets', async ({ page }) => {
    const timestamp = Date.now();
    const secretName = `e2e-test-secret-${timestamp}`;
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
    // Wait for dialog to close to ensure submission is processed
    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 30000 });

    // Verify it appears in the list
    await expect(page.getByText(secretName)).toBeVisible({ timeout: 15000 });
    await expect(page.getByText(secretKey)).toBeVisible();

    // Verify mask (24 dots) - Scope to the row
    // Find the row-like container for the secret using .group class
    const secretRow = page.locator('.group').filter({ hasText: secretName }).first();
    await expect(secretRow.getByText('••••••••••••••••••••••••')).toBeVisible();

    // Toggle visibility
    // Click the eye icon button using aria-label
    const toggleButton = secretRow.locator('button[aria-label="Show secret"]');
    await toggleButton.click();

    // Wait for the button to change to "Hide secret" to confirm state update
    await expect(secretRow.locator('button[aria-label="Hide secret"]')).toBeVisible();

    // Find the span containing the value and check text
    // The server redacts values in the list API, but after reveal we expect the real value
    await expect(secretRow.getByText(secretValue)).toBeVisible({ timeout: 5000 });

    // Delete the secret
    await secretRow.locator('button[aria-label="Delete secret"]').click();

    // Verify it's gone
    await expect(page.getByText(secretName)).not.toBeVisible();
  });

  test('should allow bulk deleting secrets', async ({ page }) => {
    const timestamp = Date.now();
    const secretsToCreate = 3;
    const baseName = `e2e-bulk-secret-${timestamp}`;

    await page.goto('/secrets');
    await expect(page.getByRole('heading', { name: 'API Keys & Secrets' })).toBeVisible();

    // Create multiple secrets manually through the UI (acts as both test and setup)
    for (let i = 0; i < secretsToCreate; i++) {
        await page.getByRole('button', { name: 'Add Secret' }).click();
        await expect(page.getByRole('dialog')).toBeVisible();

        await page.fill('#name', `${baseName}-${i}`);
        await page.fill('#key', `KEY_${i}_${timestamp}`);
        await page.fill('#value', `val-${i}-${timestamp}`);
        await page.getByRole('button', { name: 'Save Secret' }).click();
        await expect(page.getByRole('dialog')).toBeHidden({ timeout: 15000 });
        await expect(page.getByText(`${baseName}-${i}`)).toBeVisible({ timeout: 15000 });
    }

    // Select the first two secrets
    const row0 = page.locator('.group').filter({ hasText: `${baseName}-0` }).first();
    const row1 = page.locator('.group').filter({ hasText: `${baseName}-1` }).first();
    const row2 = page.locator('.group').filter({ hasText: `${baseName}-2` }).first();

    await row0.getByRole('checkbox').check();
    await row1.getByRole('checkbox').check();

    // Verify the bulk action bar appears
    await expect(page.getByText('2 selected')).toBeVisible();

    // Setup dialog handler to automatically accept the confirmation alert
    page.once('dialog', dialog => dialog.accept());

    // Click the "Delete Selected" button in the bulk action bar
    await page.getByRole('button', { name: 'Delete Selected' }).click();

    // Verify the deleted secrets are gone
    await expect(page.getByText(`${baseName}-0`)).not.toBeVisible({ timeout: 15000 });
    await expect(page.getByText(`${baseName}-1`)).not.toBeVisible();

    // Verify the unselected secret is still there
    await expect(page.getByText(`${baseName}-2`)).toBeVisible();

    // Clean up the last one
    await row2.getByRole('checkbox').check();
    page.once('dialog', dialog => dialog.accept());
    await page.getByRole('button', { name: 'Delete Selected' }).click();
    await expect(page.getByText(`${baseName}-2`)).not.toBeVisible({ timeout: 15000 });
  });
});
