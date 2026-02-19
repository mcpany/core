/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedDebugData } from './test-data';

test.describe('Credential OAuth Flow E2E', () => {
  // Use unique ID to prevent collisions
  const credentialID = 'cred-oauth-test';

  test.beforeEach(async ({ page, request }) => {
    // Increase viewport height for long forms
    await page.setViewportSize({ width: 1280, height: 1000 });
    page.on('console', msg => console.log('BROWSER LOG:', msg.text()));

    // Seed data: Clear credentials and ensure a clean state
    await seedDebugData({
        credentials: [],
        services: [] // Clear services too if needed, or keep them.
    }, request);
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

    // Point to Debug OAuth endpoints on the REAL backend
    // Since Next.js proxies /api/v1 to backend, we can use the same host or explicit backend port.
    // The browser needs to reach this URL. If we use localhost:50050, it works locally.
    // In CI, we assume backend is at localhost:50050 or reachable.
    // To be safe, we use the backend port directly as it might not be proxied if it's an external redirect.
    // Actually, the backend will redirect the browser to this URL.
    // If we use /api/v1/... it will go through Next.js proxy -> Backend -> Debug Endpoint.
    // But authorization_url is visited by the browser directly.
    // So http://localhost:50050/api/v1/debug/oauth/authorize is correct if backend is exposed there.

    const backendUrl = process.env.BACKEND_URL || 'http://localhost:50050';
    await authUrlLabel.fill(`${backendUrl}/api/v1/debug/oauth/authorize`);
    await page.getByLabel('Token URL').fill(`${backendUrl}/api/v1/debug/oauth/token`);

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

    // The flow should redirect to debug endpoint, then back to callback page, then show success
    await expect(page.getByText('Authentication Successful')).toBeVisible({ timeout: 20000 });
    await page.getByRole('button', { name: 'Continue' }).click({ force: true });

    // Use auto-retrying toHaveURL
    await expect(page).toHaveURL(/\/credentials/);
    await expect(page.getByText('Test OAuth Cred')).toBeVisible();

    // Verify "Reconnect" button appears, indicating token is present
    await expect(row.getByRole('button', { name: 'Reconnect' })).toBeVisible();
  });
});
