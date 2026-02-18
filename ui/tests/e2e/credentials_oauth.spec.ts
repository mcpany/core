/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Credential OAuth Flow E2E', () => {

  test.beforeEach(async ({ page }) => {
    // Increase viewport height for long forms
    await page.setViewportSize({ width: 1280, height: 1000 });
  });

  test('should create oauth credential and connect', async ({ page }) => {
    await page.goto('/credentials');

    await page.getByRole('button', { name: 'New Credential' }).click({ force: true });

    // Use placeholder to avoid name conflicts
    await page.getByPlaceholder('My Credential').fill('Test OAuth Cred');

    // Select Type
    await page.getByRole('combobox', { name: 'Type' }).click({ force: true });
    const oauthOption = page.getByRole('option', { name: 'OAuth 2.0' });
    await oauthOption.waitFor({ state: 'visible' });
    await oauthOption.click({ force: true });

    // Correct label: Auth URL
    const authUrlLabel = page.getByLabel('Auth URL');
    await expect(authUrlLabel).toBeVisible({ timeout: 15000 });

    await page.getByLabel('Client ID').fill('test-client-id');
    await page.getByLabel('Client Secret').fill('test-client-secret');

    // Determine hosts based on environment
    // In CI (Docker), backend sees test runner as 'ui-tests'
    // Locally, everything is localhost
    const isCI = !!process.env.CI;
    const tokenHost = isCI ? 'http://ui-tests:9999' : 'http://localhost:9999';
    // Browser (running in ui-tests container or locally) sees oauth server on localhost:9999
    const authHost = 'http://localhost:9999';

    await authUrlLabel.fill(`${authHost}/auth`);
    await page.getByLabel('Token URL').fill(`${tokenHost}/token`);

    const saveButton = page.getByRole('button', { name: 'Save', exact: true });
    await saveButton.scrollIntoViewIfNeeded();
    await saveButton.click({ force: true });

    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10000 });

    await expect(page.getByText('Test OAuth Cred')).toBeVisible();

    const row = page.locator('tr').filter({ hasText: 'Test OAuth Cred' });
    // Click direct Edit button
    await row.getByRole('button', { name: 'Edit' }).click({ force: true });

    await expect(page.getByRole('button', { name: 'Connect Account' })).toBeVisible({ timeout: 15000 });
    await page.getByRole('button', { name: 'Connect Account' }).click({ force: true });

    await expect(page.getByText('Authentication Successful')).toBeVisible({ timeout: 20000 });
    await page.getByRole('button', { name: 'Continue' }).click({ force: true });

    // Use auto-retrying toHaveURL
    await expect(page).toHaveURL(/\/credentials/);
    await expect(page.getByText('Test OAuth Cred')).toBeVisible();
  });
});
