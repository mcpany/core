/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Credential OAuth Flow E2E', () => {
  const credentialID = 'cred-123';
  const credentials: any[] = [];

  test.beforeEach(async ({ page }) => {
    // Clear credentials for each test run if needed, but here we only have one test
    credentials.length = 0; // Clear array

    page.on('console', msg => console.log('BROWSER LOG:', msg.text()));

    // Mock List Credentials
    await page.route('**/api/v1/credentials', async route => {
        if (route.request().method() === 'GET') {
            await route.fulfill({ json: { credentials } });
        } else if (route.request().method() === 'POST') {
             // Mock Create Credential
             const req = route.request().postDataJSON();
             const newCred = {
                     id: credentialID,
                     name: req.name,
                     authentication: req.authentication,
                     token: null
             };
             credentials.push(newCred);
             await route.fulfill({
                 json: newCred
             });
        } else {
            await route.continue();
        }
    });

    // Mock Update Credential (PUT) and Get Single
    await page.route(`**/api/v1/credentials/${credentialID}`, async route => {
        if (route.request().method() === 'PUT') {
            const req = route.request().postDataJSON();
            // Update in memory mock?
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
        } else if (route.request().method() === 'DELETE') {
            const idx = credentials.findIndex(c => c.id === credentialID);
            if (idx !== -1) credentials.splice(idx, 1);
            await route.fulfill({ json: {} });
        } else {
            await route.continue();
        }
    });

    // Mock Initiate OAuth
    await page.route('**/auth/oauth/initiate', async route => {
        const req = route.request().postDataJSON();
        console.log("Initiate OAuth request payload:", JSON.stringify(req));

        if (req.credential_id !== credentialID) {
            console.error(`Mismatch credentialID in initiate: received '${req.credential_id}', expected '${credentialID}'`);
            // Fail the request explicitly so UI handles error (or we see it)
            // But strict expect() crashes the route handler which causes hang.
            // Let's fulfill with error or just log and proceed for debugging?
            // Fulfilling with 400 helps UI show error.
             await route.fulfill({
                 status: 400,
                 json: { error: 'Credential ID mismatch' }
             });
             return;
        }

        await route.fulfill({
            json: {
                authorization_url: '/auth/callback?code=mock-code&state=xyz',
                state: 'xyz'
            }
        });
    });

    // Mock Provider Page
    await page.route('**/mock-auth*', async route => {
         console.log("Mock Auth hit:", route.request().url());
         const url = new URL(route.request().url());
         const state = url.searchParams.get('state');

         await route.fulfill({
             status: 302,
             headers: {
                 'Location': `/auth/callback?code=mock-code&state=${state}`
             }
         });
    });

    // Mock Callback API
    await page.route('**/auth/oauth/callback', async route => {
        const req = route.request().postDataJSON();
        console.log("Callback API received:", req);
        if (req.credential_id !== credentialID) console.error(`Mismatch credentialID: ${req.credential_id} vs ${credentialID}`);
        if (req.code !== 'mock-code') console.error(`Mismatch code: ${req.code} vs mock-code`);

        await route.fulfill({ json: { status: 'success' } });
    });
  });

  test('should create oauth credential and connect', async ({ page }) => {
    // 1. Navigate to Credentials Page
    await page.goto('/credentials');

    // 2. Open Create Dialog
    await page.getByRole('button', { name: 'New Credential' }).click();

    // 3. Fill Form
    await page.getByPlaceholder('My Credential').fill('Test OAuth Cred');
    await page.getByLabel('Type').click();
    await page.getByRole('option', { name: 'OAuth 2.0' }).click();

    // Check fields appear
    await expect(page.getByLabel('Client ID')).toBeVisible();
    await page.getByLabel('Client ID').fill('client-id');
    await page.getByLabel('Client Secret').fill('client-secret');
    await page.getByLabel('Auth URL').fill('/mock-auth');
    await page.getByLabel('Token URL').fill('/mock-token');
    await page.getByLabel('Scopes').fill('read write');

    // Wait for request
    const requestPromise = page.waitForResponse(response =>
        response.url().includes('/api/v1/credentials') && response.request().method() === 'POST'
    );
    await page.getByRole('button', { name: 'Save' }).evaluate(b => (b as HTMLElement).click());
    await requestPromise;

    // Verify toast or success state
    try {
        await expect(page.getByText('Credential created').first()).toBeVisible({ timeout: 5000 });
    } catch (e) {
        console.log("Toast missing. Content:");
        console.log(await page.content());
        throw e;
    }

    // The dialog should close automatically on success.
    // Wait for it to close?
    await expect(page.getByRole('dialog')).toBeHidden();

    // 5. Verify Credential in List
    // The list should auto-refresh or we trigger reload?
    // CredentialList `onSuccess` calls `loadCredentials`.
    // So it should show up.
    await expect(page.getByText('Test OAuth Cred')).toBeVisible();

    // 6. Click Edit to reopen dialog and connect
    await page.getByRole('button', { name: 'Edit' }).click();

    // Now we should see "Connect Account" button because initialData.id is present.
    await expect(page.getByRole('button', { name: 'Connect Account' })).toBeVisible();

    // 7. Click Connect and Handle OAuth
    await page.getByRole('button', { name: 'Connect Account' }).click();

    // 8. Verify Callback Page
    await expect(page).toHaveURL(/.*\/auth\/callback.*/).catch(async e => {
        console.log("Current URL:", page.url());
        throw e;
    });
    await expect(page.getByText('Authentication Successful')).toBeVisible();

    // 9. Return
    // The button on callback page is "Continue"
    await page.getByRole('button', { name: 'Continue' }).click();

    // It should go to credentials page
    await page.waitForURL('**/credentials');
  });
});
