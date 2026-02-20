/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { apiClient } from '@/lib/client';

test.describe('Credential OAuth Flow E2E', () => {
  // Use a unique name to avoid conflicts
  const credentialName = `Test OAuth Cred ${Date.now()}`;

  test.beforeEach(async ({ page }) => {
    // Increase viewport height for long forms
    await page.setViewportSize({ width: 1280, height: 1000 });
    page.on('console', msg => console.log('BROWSER LOG:', msg.text()));

    // External Mock for Auth Provider
    // This mocks the 3rd party provider (e.g., Google), NOT our backend.
    await page.route('http://example.com/auth*', async route => {
        const url = route.request().url();
        console.log(`Intercepted external auth: ${url}`);
        if (url.includes('response_type=code')) {
             // Redirect back to callback
             const redirectUrl = new URL(url).searchParams.get('redirect_uri') || '';
             const state = new URL(url).searchParams.get('state') || '';
             // Simulate user approving access
             // We need to perform the redirect that the provider would do.
             // The provider redirects the browser to `redirectUrl` with code and state.
             const callback = `${redirectUrl}?code=mock-code&state=${state}`;
             console.log(`Redirecting to: ${callback}`);
             await route.fulfill({
                 status: 302,
                 headers: {
                     'Location': callback
                 }
             });
        } else {
            await route.fulfill({ status: 404, body: 'Not Found' });
        }
    });

    await page.route('http://example.com/token', async route => {
        console.log('Intercepted token exchange');
        await route.fulfill({
            json: {
                access_token: 'mock-token',
                token_type: 'Bearer',
                expires_in: 3600
            }
        });
    });
  });

  test.afterEach(async () => {
      // Cleanup
      try {
          const credentials = await apiClient.listCredentials();
          const cred = credentials.find(c => c.name === credentialName);
          if (cred && cred.id) {
              await apiClient.deleteCredential(cred.id);
              console.log(`Cleaned up credential: ${credentialName}`);
          }
      } catch (e) {
          console.error('Failed to cleanup credential', e);
      }
  });

  test('should create oauth credential and connect', async ({ page }) => {
    await page.goto('/credentials');

    await page.getByRole('button', { name: 'New Credential' }).click({ force: true });

    // Use placeholder to avoid name conflicts
    await page.getByPlaceholder('My Credential').fill(credentialName);

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

    await expect(page.getByText(credentialName)).toBeVisible();

    const row = page.locator('tr').filter({ hasText: credentialName });
    // Click direct Edit button
    await row.getByRole('button', { name: 'Edit' }).click({ force: true });

    await expect(page.getByRole('button', { name: 'Connect Account' })).toBeVisible({ timeout: 15000 });
    // Wait for button to be enabled?
    await page.getByRole('button', { name: 'Connect Account' }).click({ force: true });

    // The backend should redirect to http://example.com/auth...
    // Our page.route intercepts it and redirects back to /auth/callback
    // The callback page exchanges the code (hitting http://example.com/token via backend) and closes/redirects.

    await expect(page.getByText('Authentication Successful')).toBeVisible({ timeout: 30000 });
    await page.getByRole('button', { name: 'Continue' }).click({ force: true });

    // Use auto-retrying toHaveURL
    await expect(page).toHaveURL(/\/credentials/);
    await expect(page.getByText(credentialName)).toBeVisible();
  });
});
