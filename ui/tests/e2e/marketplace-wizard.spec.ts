/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Marketplace Wizard and Service Lifecycle', () => {

  test.beforeEach(async ({ page, request }) => {
    // Seed credential for the test
    try {
      await request.post('/api/v1/credentials', {
          data: {
            id: 'cred-1',
            name: 'Test Credential',
            authentication: { apiKey: { paramName: 'Authorization', in: 0, value: { plainText: 'secret' } } }
          },
          headers: {
              'Authorization': 'Basic ' + Buffer.from('e2e-admin:password').toString('base64')
          }
      });
    } catch (e) {
      // Ignore if already exists or fails (test might fail later if critical)
      console.log("Credential seed warning:", e);
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
    await page.getByRole('option', { name: 'Filesystem' }).click();
    await page.click('button:has-text("Next")');

    // 4. Step 2: Parameters
    await expect(page.getByText('Environment Variables / Parameters').first()).toBeVisible();

    // Check for parameter input existence and edit it
    // Filesystem template has ALLOWED_PATH default "."
    const paramInput = page.locator('input[value="."]');
    await expect(paramInput).toBeVisible();
    await paramInput.fill('/tmp');

    // Add a new parameter
    await page.getByRole('button', { name: 'Add Parameter' }).click();

    // Wait for the new input to appear (should have 2 now: ALLOWED_PATH + new one)
    await expect(page.getByPlaceholder('VAR_NAME')).toHaveCount(2);

    const newKeyInput = page.getByPlaceholder('VAR_NAME').last();
    const newValueInput = page.locator('input[placeholder="Value"]').last();
    await newKeyInput.fill('MAX_CONNECTIONS');
    await newValueInput.fill('100');

    await page.click('button:has-text("Next")');

    // 5. Step 3: Webhooks
    await expect(page.getByText('Webhooks & Transformers')).toBeVisible();
    // Add a Pre-Call Webhook
    await page.getByRole('button', { name: 'Add Pre-Call Webhook' }).click();
    await page.locator('input[placeholder="https://api.example.com/webhook"]').first().fill('https://example.com/hook');
    await page.click('button:has-text("Next")');

    // 6. Step 4: Auth
    await expect(page.getByText('4. Authentication')).toBeVisible();
    // Filesystem has no Upstream Auth, so just click Next
    await page.click('button:has-text("Next")');

    // 7. Step 5: Review
    await expect(page.getByText('Review & Finish')).toBeVisible(); // Title is "5. Review & Finish" in create-config-wizard.tsx
    // Check if JSON contains our changes
    await expect(page.getByText('"MAX_CONNECTIONS"')).toBeVisible();
    await expect(page.getByText('"100"')).toBeVisible();
    await expect(page.getByText('/tmp')).toBeVisible();

    await page.click('button:has-text("Finish & Save")');

    // 8. Verify Saved to Local
    await expect(page.getByRole('tab', { name: 'Local' })).toBeVisible();
    await page.getByRole('tab', { name: 'Local' }).click();

    // 9. Instantiate
    await expect(page.getByRole('button', { name: 'Instantiate' }).first()).toBeVisible();
    await page.getByRole('button', { name: 'Instantiate' }).first().click();

    await expect(page.getByRole('dialog', { name: 'Instantiate Service' })).toBeVisible();
    const uniqueName = `filesystem-test-${Date.now()}`;
    const nameInput = page.locator('#service-name-input');
    await expect(nameInput).toBeVisible();
    await nameInput.fill(uniqueName);

    // Mock the register service call
    const registerPromise = page.waitForResponse(response =>
        response.url().includes('/api/v1/services') && response.status() === 200
    );

    await page.click('button:has-text("Create Instance")');
    await registerPromise;

    // Verify toast or closing of dialog
    await expect(page.getByRole('dialog', { name: 'Instantiate Service' })).toBeHidden();
  });
});
