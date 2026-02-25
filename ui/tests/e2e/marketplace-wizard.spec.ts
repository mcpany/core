/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect, request } from '@playwright/test';

test.describe('Marketplace Wizard and Service Lifecycle', () => {

  test.beforeEach(async ({ page }) => {
    // Seed credential via API (Real Backend)
    const apiContext = await request.newContext();
    // Ensure we have the credential the test expects
    // Note: If backend persists state, this might conflict if id exists.
    // But test environment usually resets or uses ephemeral DB.
    // If running against persistent dev server, we should handle error or use unique ID.
    // E2E usually runs in fresh container.
    try {
        await apiContext.post('/api/v1/credentials', {
            data: {
                id: 'cred-1',
                name: 'Test Credential',
                authentication: { apiKey: { paramName: 'Authorization', in: 0, value: { plainText: 'secret' } } }
            }
        });
    } catch (e) {
        // Ignore if exists? Or check response.
        console.log("Credential creation attempted");
    }
  });

  test('Complete CUJ: Create Config -> Instantiate -> Manage', async ({ page }) => {
    // 1. Navigate to Marketplace
    await page.goto('/marketplace');
    await expect(page.getByText('Marketplace', { exact: true }).first()).toBeVisible();

    // 2. Open Wizard
    await page.getByRole('button', { name: 'Create Config' }).click();
    await expect(page.getByRole('dialog', { name: 'Create Upstream Service' })).toBeVisible();

    // 3. Step 1: Service Type
    await expect(page.getByText('Service Type')).toBeVisible();
    await page.getByRole('combobox').click();
    // Select "Test Service" which we seeded in backend
    await page.getByRole('option', { name: 'Test Service' }).click();
    await page.click('button:has-text("Next")');

    // 4. Step 2: Parameters
    await expect(page.getByText('Environment Variables / Parameters').first()).toBeVisible();

    // Check for parameter input existence and edit it
    // "Test Service" has TEST_VAR with default "default"
    const paramInput = page.locator('input[value="default"]');
    await expect(paramInput).toBeVisible();
    await paramInput.fill('modified-value');

    // Add a new parameter
    await page.getByRole('button', { name: 'Add Parameter' }).click();

    // Wait for the new input to appear (should have 2 now: TEST_VAR + new one)
    await expect(page.getByPlaceholder('VAR_NAME')).toHaveCount(2);

    const newKeyInput = page.getByPlaceholder('VAR_NAME').last();
    const newValueInput = page.locator('input[placeholder="Value"]').last();
    await newKeyInput.fill('EXTRA_VAR');
    await newValueInput.fill('extra-val');

    await page.click('button:has-text("Next")');

    // 5. Step 3: Webhooks
    await expect(page.getByText('Webhooks & Transformers')).toBeVisible();
    // Add a Pre-Call Webhook
    await page.getByRole('button', { name: 'Add Pre-Call Webhook' }).click();
    await page.locator('input[placeholder="https://api.example.com/webhook"]').first().fill('https://example.com/hook');
    await page.click('button:has-text("Next")');

    // 6. Step 4: Auth
    // "Test Service" has API Key Auth configured, so this step should appear.
    await expect(page.getByText('4. Authentication')).toBeVisible();
    // Verify "Test Only" alert is present
    await expect(page.getByText('Test Connection Only')).toBeVisible();

    // Verify we can see the credential we seeded
    await page.getByRole('combobox').click({ force: true });
    await expect(page.getByRole('option', { name: 'Test Credential' })).toBeVisible({ timeout: 10000 });
    // Select Test Credential
    await page.getByRole('option', { name: 'Test Credential' }).click();

    // Helper: Test Connection
    // Note: The real "Test Service" command is `echo hello`. It ignores Auth header.
    // But `auth-test` endpoint in backend should succeed if credential is valid and connection works?
    // Actually `auth-test` tries to make a request.
    // `echo` command (CLI) doesn't support Auth Check in the same way as HTTP.
    // But wizard might skip verification or it might pass if command runs.
    // If it's CLI, Auth Test might not apply or might just check if credential exists.
    // Let's assume we skip clicking "Test Connection" button if it's potentially flaky with CLI,
    // or just proceed. The previous test clicked it.
    // Let's try clicking it.
    await page.getByRole('button', { name: 'Test Connection' }).click();
    // Expect success message (toast or alert or status)
    // If it fails (because CLI doesn't support auth test?), we might need to adjust.
    // For now, let's wait for it.
    await expect(page.getByText('Connection verification successful')).toBeVisible({ timeout: 60000 });

    await page.click('button:has-text("Next")');

    // 7. Step 5: Review
    await expect(page.getByText('Review & Finish')).toBeVisible(); // Title is "5. Review & Finish" in create-config-wizard.tsx
    // Check if JSON contains our changes
    await expect(page.getByText('"EXTRA_VAR"')).toBeVisible();
    await expect(page.getByText('"extra-val"')).toBeVisible();
    await expect(page.getByText('"modified-value"')).toBeVisible();

    await page.click('button:has-text("Finish & Save")');

    // 8. Verify Saved to Local
    await expect(page.getByRole('tab', { name: 'Local' })).toBeVisible();
    await page.getByRole('tab', { name: 'Local' }).click();

    // 9. Instantiate
    // We should see the new config in the list.
    // It might be named "Test Service" or similar.
    // Find the instantiate button for it.
    await expect(page.getByRole('button', { name: 'Instantiate' }).first()).toBeVisible();
    await page.getByRole('button', { name: 'Instantiate' }).first().click();

    await expect(page.getByRole('dialog', { name: 'Instantiate Service' })).toBeVisible();
    const uniqueName = `test-service-instance-${Date.now()}`;
    const nameInput = page.locator('#service-name-input');
    await expect(nameInput).toBeVisible();
    await nameInput.fill(uniqueName);

    // Mock the register service call? NO, use real backend.
    // The previous test mocked:
    // const registerPromise = page.waitForResponse(response =>
    //     response.url().includes('/api/v1/services') && response.status() === 200
    // );
    // We can still wait for the response to confirm it happened.
    const registerPromise = page.waitForResponse(response =>
        response.url().includes('/api/v1/services') && response.request().method() === 'POST' && response.status() === 200
    );

    await page.click('button:has-text("Create Instance")');
    await registerPromise;

    // Verify toast or closing of dialog
    await expect(page.getByRole('dialog', { name: 'Instantiate Service' })).toBeHidden();
  });
});
