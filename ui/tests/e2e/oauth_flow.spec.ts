/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { createServer, Server } from 'http';
import { AddressInfo } from 'net';

test.describe('OAuth Flow Integration', () => {
  let authServer: Server;
  let authPort: number;
  let authUrl: string;

  // Setup local "Provider" server
  test.beforeAll(async () => {
    return new Promise<void>((resolve) => {
        authServer = createServer((req, res) => {
            const url = new URL(req.url || '', `http://localhost:${authPort}`);
            console.log(`Auth Server received: ${req.method} ${req.url}`);

            res.setHeader('Access-Control-Allow-Origin', '*'); // If needed

            if (url.pathname === '/auth') {
                // 1. Authorization Endpoint
                // Redirect back to the redirect_uri with a code
                const redirectUri = url.searchParams.get('redirect_uri');
                const state = url.searchParams.get('state');
                if (redirectUri) {
                    res.statusCode = 302;
                    res.setHeader('Location', `${redirectUri}?code=mock_code&state=${state}`);
                    res.end();
                } else {
                    res.statusCode = 400;
                    res.end('Missing redirect_uri');
                }
            } else if (url.pathname === '/token') {
                // 2. Token Endpoint
                // Return valid JSON
                if (req.method === 'POST') {
                    res.statusCode = 200;
                    res.setHeader('Content-Type', 'application/json');
                    res.end(JSON.stringify({
                        access_token: 'mock-provider-token',
                        token_type: 'Bearer',
                        scope: 'read:user'
                    }));
                } else {
                     res.statusCode = 405;
                     res.end();
                }
            } else {
                res.statusCode = 404;
                res.end();
            }
        });

        authServer.listen(0, '0.0.0.0', () => {
            authPort = (authServer.address() as AddressInfo).port;
            // Use 'localhost' as it works for Browser running in same context (Host or Container)
            // and we are mocking the Backend callback, so Backend reachability is not required.
            authUrl = `http://localhost:${authPort}`;
            console.log(`Auth Provider started on port ${authPort} at ${authUrl}`);
            resolve();
        });
    });
  });

  test.afterAll(async () => {
    return new Promise<void>((resolve) => {
        authServer.close(() => resolve());
    });
  });

  test('should complete the OAuth flow via Auth Wizard', async ({ page }) => {
    // 1. Create Credential via UI pointing to our local server
    // We use the Wizard to create a NEW configuration which creates the credential implicitly?
    // No, the test flow says: "Create Config" -> "Select Credential" -> "Connect with OAuth".
    // This implies we need to CREATE a credential first or the Wizard lets us create one?
    // The previous test mocked `credentials` list to include one.
    // The Wizard likely has a "Add Credential" flow or we should create it beforehand in "Credentials" page.

    // Let's go to Credentials page first to create the OAuth credential using our local server details.
    await page.goto('/credentials');
    const credName = `OAuth Test Cred ${Date.now()}`;

    await page.getByRole('button', { name: 'New Credential' }).click();
    await page.getByPlaceholder('My Credential').fill(credName);

    // Select OAuth2 type if available (Dropdown?)
    // Assuming default is API Key, we might need to switch type.
    // Checking previous test content: it defined `authentication: { oauth2: ... }`.
    // The UI must support creating OAuth2 credentials.
    // Let's check if there is a type selector.
    // If not visible in previous test, maybe it was just filling fields?
    // Wait, the previous test MOCKED the credential. It didn't create it via UI.
    // We MUST create it via UI to be E2E.

    // If the UI has a type selector:
    // If the UI has a type selector (it shows current type as text in combobox)
    // We target the combobox directly.
    const typeSelector = page.getByRole('combobox', { name: 'Type' }); // Or just 'combobox' if unique
    // But name 'Type' might be inferred from label.
    // If not sure, use getByLabel('Type').

    if (await typeSelector.isVisible()) {
        await typeSelector.click();
        await page.getByRole('option', { name: 'OAuth 2.0' }).click();
        // Wait for selection to be applied
        await expect(typeSelector).toHaveText('OAuth 2.0');
    }

    // Fill OAuth fields
    await page.getByLabel('Client ID').fill('client-id');
    await page.getByLabel('Client Secret').fill('client-secret');
    await page.getByLabel('Auth URL').fill(`${authUrl}/auth`);
    await page.getByLabel('Token URL').fill(`${authUrl}/token`);
    await page.getByLabel('Scopes').fill('read:user');

    await page.getByLabel('Scopes').press('Enter');
    // await page.getByRole('button', { name: 'Save' }).click({ force: true });
    await expect(page.getByText(credName)).toBeVisible();

    // Now go to Marketplace to use it
    await page.goto('/marketplace');
    await page.getByRole('button', { name: 'Create Config', exact: true }).click({ force: true });

    // Step 1: Type & Template
    await page.getByPlaceholder('e.g. My Postgres DB').fill('OAuth Test Service');
    await page.getByRole('button', { name: 'Next', exact: true }).click({ force: true });

    // Step 2: Parameters (Skip)
    await page.getByRole('button', { name: 'Next', exact: true }).click({ force: true });

    // Step 3: Webhooks (Skip)
    await page.getByRole('button', { name: 'Next', exact: true }).click({ force: true });

    // Step 4: Authentication
    await expect(page.getByText('Select Credential for Testing')).toBeVisible({ timeout: 15000 });

    // Select our credential
    await page.getByRole('combobox').click({ force: true });
    await page.getByRole('option', { name: credName }).click({ force: true });

    // Click Connect
    // This will open a popup or redirect. Playwright handles popups if expected.
    // If it redirects the main page, we just wait.
    await page.getByRole('button', { name: 'Connect with OAuth' }).click({ force: true });

    // Flow:
    // 1. App redirects to authUrl/auth
    // 2. authUrl/auth redirects to App Callback
    // 3. App Callback verifies and shows Success

    await expect(page.getByText('Authentication Successful')).toBeVisible({ timeout: 30000 });

    await page.getByRole('button', { name: 'Continue' }).click({ force: true });

    // BACK IN WIZARD
    await expect(page.getByText('Account Connected')).toBeVisible({ timeout: 20000 });

    await page.getByRole('button', { name: 'Next', exact: true }).click({ force: true }); // Review
    await page.getByRole('button', { name: /Finish/ }).click({ force: true });

    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10000 });

    // Cleanup: Delete Credential
    await page.goto('/credentials');
    const row = page.locator('tr').filter({ hasText: credName });
    page.on('dialog', d => d.accept());
    await row.getByRole('button', { name: 'Delete' }).click();
    // Handle custom dialog if needed
    try {
        const dialog = page.getByRole('alertdialog');
        if (await dialog.isVisible({ timeout: 2000 })) {
             await dialog.getByRole('button', { name: 'Delete' }).click();
        }
    } catch {}
    await expect(page.getByText(credName)).toBeHidden();
  });
});
