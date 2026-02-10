/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('OAuth Flow Integration', () => {
  const credentialID = 'cred-oauth-1';
  let callbackCalled = false;
  const credentials: any[] = [
    {
      id: credentialID,
      name: 'GitHub OAuth',
      authentication: {
        oauth2: {
          clientId: { value: { plainText: 'client-id' } },
          authorizationUrl: 'http://127.0.0.1:38817/auth',
          tokenUrl: 'http://127.0.0.1:38817/token',
          scopes: 'read:user'
        }
      },
      token: null
    }
  ];

  test.beforeEach(async ({ page }) => {
    // Increase viewport height for long forms/lists
    await page.setViewportSize({ width: 1280, height: 1000 });

    callbackCalled = false;
    // Reset credentials for each test if multiple tests existed
    credentials[0].token = null;

    page.on('console', msg => console.log('BROWSER LOG:', msg.text()));

    await page.route('**/api/v1/credentials', async route => {
      console.log(`Mocking list credentials, token: ${!!credentials[0].token}`);
      await route.fulfill({
        json: { credentials }
      });
    });

    await page.route((url) => url.pathname.includes('/auth/oauth/'), async route => {
      const urlStr = route.request().url();
      if (urlStr.includes('/initiate')) {
        const origin = new URL(page.url()).origin;
        await route.fulfill({
          json: {
            authorization_url: `${origin}/auth/callback?code=mock_code&state=test_state_123`,
            state: 'test_state_123'
          }
        });
      } else if (urlStr.includes('/callback')) {
        callbackCalled = true;
        // UPDATE credentials to have a token
        credentials[0].token = { accessToken: 'mock-token' };
        await route.fulfill({ json: { status: 'success' } });
      } else {
        await route.continue();
      }
    });

    // Mock service create
    await page.route('**/api/v1/services', async route => {
        if (route.request().method() === 'POST') {
             await route.fulfill({ json: { id: 'test-service' } });
        } else {
            await route.continue();
        }
    });

    // Mock templates list for marketplace
    await page.route('**/api/v1/registration/templates', async route => {
        await route.fulfill({ json: { templates: [] } });
    });
  });

  test('should complete the OAuth flow via Auth Wizard', async ({ page }) => {
    await page.goto('/marketplace');
    await page.getByRole('button', { name: 'Create Config', exact: true }).click({ force: true });

    // Step 1: Type & Template in Wizard
    // The wizard shows "Manual / Custom" in the template selection card by default
    await page.getByPlaceholder('e.g. My Postgres DB').fill('OAuth Test Service');
    await page.getByRole('button', { name: 'Next', exact: true }).click({ force: true });

    // Step 2: Parameters (Skip or click Next)
    await page.getByRole('button', { name: 'Next', exact: true }).click({ force: true });

    // Step 3: Webhooks (Skip or click Next)
    await page.getByRole('button', { name: 'Next', exact: true }).click({ force: true });

    // Step 4: Authentication
    await expect(page.getByText('Select Credential for Testing')).toBeVisible({ timeout: 15 * 1000 });

    await page.getByRole('combobox').click({ force: true });
    await page.getByRole('option', { name: 'GitHub OAuth' }).click({ force: true });

    await page.getByRole('button', { name: 'Connect with OAuth' }).click({ force: true });

    // Success check in callback page
    await expect(page.getByText('Authentication Successful')).toBeVisible({ timeout: 30 * 1000 });
    expect(callbackCalled).toBeTruthy();

    await page.getByRole('button', { name: 'Continue' }).click({ force: true });

    // BACK IN WIZARD
    // Now it should show Account Connected because we updated the mock
    await expect(page.getByText('Account Connected')).toBeVisible({ timeout: 20 * 1000 });

    await page.getByRole('button', { name: 'Next', exact: true }).click({ force: true }); // Go to Review
    await page.getByRole('button', { name: /Finish/ }).click({ force: true });

    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10 * 1000 });
  });
});
