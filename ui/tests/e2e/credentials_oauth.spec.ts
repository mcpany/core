/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Credential OAuth Flow E2E', () => {
  const credentialID = 'cred-123';
  const credentials: any[] = [];

  test.beforeEach(async ({ page }) => {
    // Increase viewport height for long forms
    await page.setViewportSize({ width: 1280, height: 1000 });

    credentials.length = 0;
    page.on('console', msg => console.log('BROWSER LOG:', msg.text()));

    await page.route('**/api/v1/credentials', async route => {
        if (route.request().method() === 'GET') {
            await route.fulfill({ json: { credentials } });
        } else if (route.request().method() === 'POST') {
             const req = route.request().postDataJSON();
             const newCred = {
                     id: credentialID,
                     name: req.name,
                     authentication: req.authentication,
                     token: null
             };
             credentials.push(newCred);
             await route.fulfill({ json: newCred });
        } else {
            await route.continue();
        }
    });

    await page.route(`**/api/v1/credentials/${credentialID}`, async route => {
        const method = route.request().method();
        if (method === 'PUT') {
            const req = route.request().postDataJSON();
            const cred = credentials.find(c => c.id === credentialID);
            if (cred) {
                cred.name = req.name;
                cred.authentication = req.authentication;
            }
             await route.fulfill({
                 json: {
                     id: credentialID,
                     name: req.name,
                     authentication: req.authentication,
                     token: req.token
                 }
             });
        } else if (method === 'DELETE') {
            const idx = credentials.findIndex(c => c.id === credentialID);
            if (idx !== -1) credentials.splice(idx, 1);
            await route.fulfill({ json: {} });
        } else {
            await route.continue();
        }
    });

    await page.route((url) => url.pathname.includes('/auth/oauth/'), async route => {
        const urlStr = route.request().url();
        console.log(`OAuth mock hit for ${urlStr}`);
        if (urlStr.includes('/initiate')) {
            const origin = new URL(page.url()).origin;
            await route.fulfill({
                json: {
                    authorization_url: `${origin}/auth/callback?code=mock-code&state=xyz`,
                    state: 'xyz'
                }
            });
        } else if (urlStr.includes('/callback')) {
            // Find credential and update token
            const cred = credentials.find(c => c.id === credentialID);
            if (cred) cred.token = { accessToken: 'mock-token' };
            await route.fulfill({ json: { status: 'success' } });
        } else {
            await route.continue();
        }
    });
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
    await authUrlLabel.fill('http://example.com/auth');
    await page.getByLabel('Token URL').fill('http://example.com/token');

    const saveButton = page.getByRole('button', { name: 'Save', exact: true });
    await saveButton.scrollIntoViewIfNeeded();
    await saveButton.click({ force: true });

    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10000 });

    await expect(page.getByText('Test OAuth Cred')).toBeVisible();

    const row = page.locator('tr').filter({ hasText: 'Test OAuth Cred' });
    // Click direct Edit button
    await row.getByRole('button', { name: 'Edit' }).click();

    await expect(page.getByRole('button', { name: 'Connect Account' })).toBeVisible({ timeout: 15000 });
    await expect(page.getByRole('button', { name: 'Connect Account' })).toBeEnabled();
    await page.getByRole('button', { name: 'Connect Account' }).click();

    await expect(page.getByText('Authentication Successful')).toBeVisible({ timeout: 20000 });
    await page.getByRole('button', { name: 'Continue' }).click({ force: true });

    // Use auto-retrying toHaveURL
    await expect(page).toHaveURL(/\/credentials/);
    await expect(page.getByText('Test OAuth Cred')).toBeVisible();
  });
});
