/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Marketplace Wizard and Service Lifecycle', () => {

  test.beforeEach(async ({ page, request }) => {
    // Seed credential via API
    // Ensure we are authenticated if needed (server default setup might require it or not)
    // docker-compose env MCPANY_API_KEY=test-token
    const creds = await request.post('/api/v1/credentials', {
        headers: {
            'X-API-Key': 'test-token',
            'Content-Type': 'application/json'
        },
        data: {
          id: 'cred-1',
          name: 'Test Credential',
          authentication: { apiKey: { paramName: 'Authorization', in: 0, value: { plainText: 'secret' } } }
        },
        ignoreHTTPSErrors: true
    });
    // If credential already exists (conflict), that's fine for subsequent runs if persistence is on,
    // but tests usually run ephemeral.
    if (!creds.ok()) {
        console.log(`Credential creation status: ${creds.status()} ${await creds.text()}`);
    }
    // expect(creds.ok()).toBeTruthy(); // Optional: assert if strict
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
    // Use seeded "PostgreSQL" template
    await page.getByRole('option', { name: 'PostgreSQL', exact: true }).click();
    await page.click('button:has-text("Next")');

    // 4. Step 2: Parameters
    await expect(page.getByText('Environment Variables / Parameters').first()).toBeVisible();

    // Check for parameter input existence and edit it
    // Default value in seeds.go is "postgresql://postgres:postgres@localhost:5432/postgres"
    const paramInput = page.locator('input[value="postgresql://postgres:postgres@localhost:5432/postgres"]');
    await expect(paramInput).toBeVisible();
    // Update to point to the real postgres service in docker-compose
    await paramInput.fill('postgresql://test:test@postgres:5432/testdb');

    // Add a new parameter
    await page.getByRole('button', { name: 'Add Parameter' }).click();

    // Wait for the new input to appear
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
    // Verify "Test Only" alert is present
    await expect(page.getByText('Test Connection Only')).toBeVisible();

    // Verify we can see the credential we seeded
    await page.getByRole('combobox').click({ force: true });
    await expect(page.getByRole('option', { name: 'Test Credential' })).toBeVisible({ timeout: 10000 });
    // Select Test Credential
    await page.getByRole('option', { name: 'Test Credential' }).click();

    // Helper: Test Connection
    await page.getByRole('button', { name: 'Test Connection' }).click();
    // Expect success message
    await expect(page.getByText('Connection verification successful')).toBeVisible({ timeout: 60000 });

    await page.click('button:has-text("Next")');

    // 7. Step 5: Review
    await expect(page.getByText('Review & Finish')).toBeVisible();
    // Check if JSON contains our changes
    await expect(page.getByText('"MAX_CONNECTIONS"')).toBeVisible();
    await expect(page.getByText('"100"')).toBeVisible();
    await expect(page.getByText('postgresql://test:test@postgres:5432/testdb')).toBeVisible();

    await page.click('button:has-text("Finish & Save")');

    // 8. Verify Saved to Local
    await expect(page.getByRole('tab', { name: 'Local' })).toBeVisible();
    await page.getByRole('tab', { name: 'Local' }).click();

    // 9. Instantiate
    await expect(page.getByRole('button', { name: 'Instantiate' }).first()).toBeVisible();
    await page.getByRole('button', { name: 'Instantiate' }).first().click();

    await expect(page.getByRole('dialog', { name: 'Instantiate Service' })).toBeVisible();
    const uniqueName = `postgres-test-${Date.now()}`;
    const nameInput = page.locator('#service-name-input');
    await expect(nameInput).toBeVisible();
    await nameInput.fill(uniqueName);

    // No need to mock register service call, let it hit the backend
    const registerPromise = page.waitForResponse(response =>
        response.url().includes('/api/v1/services') && response.status() === 200
    );

    await page.click('button:has-text("Create Instance")');
    await registerPromise;

    // Verify toast or closing of dialog
    await expect(page.getByRole('dialog', { name: 'Instantiate Service' })).toBeHidden();
  });
});
