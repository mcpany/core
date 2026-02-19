/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { createServer, Server } from 'http';
import { AddressInfo } from 'net';

test.describe('Credential OAuth Flow E2E', () => {
  let mockProviderServer: Server;
  let mockProviderUrl: string;

  test.beforeAll(async () => {
    // Start a mock OAuth provider server accessible by the backend
    mockProviderServer = createServer((req, res) => {
      // Enable CORS
      res.setHeader('Access-Control-Allow-Origin', '*');

      if (req.url?.startsWith('/auth')) {
        // Extract redirect_uri and state
        const url = new URL(req.url, `http://${req.headers.host}`);
        const redirectUri = url.searchParams.get('redirect_uri');
        const state = url.searchParams.get('state');

        if (redirectUri && state) {
          // Redirect back to the application callback
          res.writeHead(302, {
            'Location': `${redirectUri}?code=test-auth-code&state=${state}`
          });
          res.end();
        } else {
          res.writeHead(400);
          res.end('Missing params');
        }
      } else if (req.url?.startsWith('/token')) {
        // Return JSON access token
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({
          access_token: 'mock-provider-access-token',
          token_type: 'Bearer',
          expires_in: 3600
        }));
      } else {
        res.writeHead(404);
        res.end();
      }
    });

    await new Promise<void>((resolve) => {
      mockProviderServer.listen(0, () => {
        const port = (mockProviderServer.address() as AddressInfo).port;
        // Use 127.0.0.1 to avoid IPv6 issues if localhost resolves ambiguously
        mockProviderUrl = `http://127.0.0.1:${port}`;
        console.log(`Mock OAuth Provider running at ${mockProviderUrl}`);
        resolve();
      });
    });
  });

  test.afterAll(async () => {
    await new Promise<void>(resolve => mockProviderServer.close(() => resolve()));
  });

  test.beforeEach(async ({ page, request }) => {
    // Increase viewport height for long forms
    await page.setViewportSize({ width: 1280, height: 1000 });
    page.on('console', msg => console.log('BROWSER LOG:', msg.text()));

    // Seed/Reset database via Backend API to ensure clean state
    // We send an empty list to clear credentials, or we could leave it
    // and rely on unique names. For robustness, let's just ensure we don't conflict.
    // Ideally we would delete the specific credential we are about to create if it exists.
    // The debug/seed endpoint upserts.

    // For this test, we create via UI, so we just want to ensure the backend is responsive.
    // We don't mock /api/v1/credentials anymore.
  });

  test('should create oauth credential and connect', async ({ page }) => {
    await page.goto('/credentials');

    await page.getByRole('button', { name: 'New Credential' }).click({ force: true });

    // Use placeholder to avoid name conflicts
    const credName = `Test OAuth Cred ${Date.now()}`;
    await page.getByPlaceholder('My Credential').fill(credName);

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

    // Use our local mock provider URL
    await authUrlLabel.fill(`${mockProviderUrl}/auth`);
    await page.getByLabel('Token URL').fill(`${mockProviderUrl}/token`);

    const saveButton = page.getByRole('button', { name: 'Save', exact: true });
    await saveButton.scrollIntoViewIfNeeded();
    await saveButton.click({ force: true });

    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10000 });

    await expect(page.getByText(credName)).toBeVisible();

    const row = page.locator('tr').filter({ hasText: credName });
    // Click direct Edit button
    await row.getByRole('button', { name: 'Edit' }).click({ force: true });

    await expect(page.getByRole('button', { name: 'Connect Account' })).toBeVisible({ timeout: 15000 });

    // This will trigger navigation to mockProviderUrl/auth -> redirect to /oauth/callback -> calls backend -> backend calls mockProviderUrl/token
    await page.getByRole('button', { name: 'Connect Account' }).click({ force: true });

    await expect(page.getByText('Authentication Successful')).toBeVisible({ timeout: 20000 });
    await page.getByRole('button', { name: 'Continue' }).click({ force: true });

    // Use auto-retrying toHaveURL
    await expect(page).toHaveURL(/\/credentials/);
    await expect(page.getByText(credName)).toBeVisible();
  });
});
