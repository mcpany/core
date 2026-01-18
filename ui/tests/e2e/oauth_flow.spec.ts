/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import http from 'http';
import { AddressInfo } from 'net';

test.describe('OAuth Flow Integration', () => {
  let mockProviderFn: http.Server;
  let mockProviderUrl: string;

  test.beforeAll(async () => {
    // Start a simple HTTP server to act as OAuth Provider
    mockProviderFn = http.createServer((req, res) => {
      const url = new URL(req.url || '', `http://${req.headers.host}`);
      if (url.pathname === '/auth') {
        const redirectUri = url.searchParams.get('redirect_uri');
        const state = url.searchParams.get('state');
        if (redirectUri && state) {
            // Simulate user approval by redirecting back immediately
            res.writeHead(302, {
                'Location': `${redirectUri}?code=mock_code&state=${state}`
            });
            res.end();
            return;
        }
      }
      res.writeHead(404);
      res.end();
    });

    await new Promise<void>((resolve) => {
        mockProviderFn.listen(0, '127.0.0.1', () => resolve());
    });
    const addr = mockProviderFn.address() as AddressInfo;
    mockProviderUrl = `http://127.0.0.1:${addr.port}`;
    console.log(`Mock OAuth Provider running at ${mockProviderUrl}`);
  });

  test.afterAll(async () => {
    mockProviderFn.close();
  });

  test('should complete the OAuth flow via Auth Wizard', async ({ page }) => {
    // 1. Mock Credentials List to return an OAuth Credential
    const oauthCred = {
        id: 'cred-oauth-1',
        name: 'GitHub OAuth',
        authentication: {
            oauth2: {
                clientId: { value: { plainText: 'client-id' } },
                authorizationUrl: `${mockProviderUrl}/auth`,
                tokenUrl: `${mockProviderUrl}/token`,
                scopes: 'read:user'
            }
        }
    };

    await page.route('**/api/v1/credentials', async route => {
        if (route.request().method() === 'GET') {
            await route.fulfill({ json: [oauthCred] });
        } else {
            await route.continue();
        }
    });

    // 2. Mock Initiate OAuth API
    // The UI calls this to get the Authorization URL
    await page.route('**/auth/oauth/initiate', async route => {
        const body = route.request().postDataJSON();
        // Expect valid body
        if (body.credential_id === 'cred-oauth-1' && body.redirect_url) {
            await route.fulfill({
                json: {
                    authorization_url: `${mockProviderUrl}/auth?state=test_state_123&redirect_uri=${encodeURIComponent(body.redirect_url)}`,
                    state: 'test_state_123'
                }
            });
        } else {
            await route.fulfill({ status: 400, body: 'Invalid request' });
        }
    });

    // 3. Mock Callback API
    // The UI calls this after returning from provider
    let callbackCalled = false;
    await page.route('**/auth/oauth/callback', async route => {
        callbackCalled = true;
        const body = route.request().postDataJSON();
        // Expect code and state match
        if (body.code === 'mock_code') {
            await route.fulfill({ json: { status: 'success' } });
        } else {
            await route.fulfill({ status: 400, body: 'Invalid code' });
        }
    });

    // 4. Start Flow
    // We navigate to wizard usually via New Service -> select Template or blank.
    // For this test, we can try to test the Wizard components in isolation OR go through the flow.
    // Let's assume we can go to /marketplace and click something, or direct to a wizard route if it exists.
    // Currently wizard is a modal or a route?
    // `app/marketplace/page.tsx` starts the wizard.
    // Let's go to /marketplace.

    // Pass empty services/templates list to avoid noise
    await page.route('**/api/v1/services', async route => route.fulfill({ json: [] }));
    await page.route('**/api/v1/templates', async route => route.fulfill({ json: [] }));

    await page.goto('/marketplace');

    // Click "Configure" on a "Blank Service" or similar if we can find one.
    // Or simpler: The wizard seems to be triggered by "Add Service" or selecting a template.
    // Let's look for a button to start wizard.
    // Assuming "Create Service" or similar exists.
    // If not, we might need to verify marketplace page content.
    // Ignoring the exact entry point, let's assume we are IN the wizard.
    // Can we shallow mount? No, E2E.
    // Let's click "Create from Scratch" if available.

    await expect(page.getByRole('button', { name: 'Create Config' })).toBeVisible();
    await page.getByRole('button', { name: 'Create Config' }).click();

    // Now in Wizard.
    // Step 1: Service Type
    // Assuming Step 1 is "Parameters" or "Service Type".
    // If it's Service Type (step-service-type.tsx), we might need to select something.
    // But let's see if we can just navigate or if we are already at a step.
    // `CreateConfigWizard` defaults to step 0.
    // If step 0 is `StepServiceType`, we need to pick one?
    // Let's assume we can click Next if default is selected?
    // Or we might need to fill "Service Name" which is usually in parameters.
    // Wait for Wizard Dialog
    await expect(page.getByRole('dialog', { name: 'Create Upstream Service Config' })).toBeVisible();

    // Step 1: Service Type (if applicable) or Parameters.
    // Let's assume we need to click Next until we hit Auth.
    // Or check what step we are on.
    // We can just try to click Next.

    // If Step 1 is Service Type:
    if (await page.getByText('1. Select Service Type').isVisible()) {
        await page.getByLabel('Service Name').fill('My OAuth Service');

        // Ensure we select a template to populate config (trigger change)
        await page.getByLabel('Template').click();
        await page.getByRole('option', { name: 'Filesystem' }).click();

        // Wait for state update
        await page.waitForTimeout(1000);

        await page.getByRole('button', { name: 'Next', exact: true }).click();

         // Check for validation error (robust)
        if (await page.getByText('Validation Error').first().isVisible()) {
            // ...
        }
    }

    // Step 2: Configure Parameters
    await expect(page.getByText('2. Configure Parameters')).toBeVisible();
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // Step 3: Webhooks
    await expect(page.getByText('3. Webhooks & Transformers')).toBeVisible();
    await page.getByRole('button', { name: 'Next', exact: true }).click();

     // Step 4: Authentication
    await expect(page.getByText('4. Authentication')).toBeVisible();

     // Step 4: Authentication
    await expect(page.getByText('4. Authentication')).toBeVisible();

    // Step 2: Connection / Auth?
    // If Auth is step 2.
    // Wait for "Select Credential for Testing".
    await expect(page.getByText('Select Credential for Testing')).toBeVisible();

    // Select our credential
    await page.getByRole('combobox').click();
    await page.getByRole('option', { name: 'GitHub OAuth' }).click();

    // Verify "Connect with OAuth" button appears
    const connectBtn = page.getByRole('button', { name: 'Connect with OAuth' });
    await expect(connectBtn).toBeVisible();

    // Click it
    await connectBtn.click();

    // This should redirect to external mock provider, then back to /auth/callback
    // We wait for the URL to change to /auth/callback
    await page.waitForURL(/.*\/auth\/callback.*/);

    // Verify Success UI
    await expect(page.getByText('Authentication Successful')).toBeVisible();
    await expect(page.getByText('You have successfully connected your account.')).toBeVisible();

    // Verify API called
    expect(callbackCalled).toBe(true);

    // Click Continue
    await page.getByRole('button', { name: 'Continue' }).click();

    // Wizard state is lost on redirect, so we expect to be back at Marketplace page
    await expect(page.getByRole('heading', { name: 'Marketplace' })).toBeVisible();
  });
});
