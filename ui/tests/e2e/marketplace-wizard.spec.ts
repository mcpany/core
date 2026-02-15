/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Marketplace Wizard and Service Lifecycle', () => {

  test.beforeEach(async ({ page, request }) => {
    // Seed real backend with data
    await request.post('/api/v1/debug/seed', {
      headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' },
      data: {
        credentials: [
          {
            id: 'cred-1',
            name: 'Test Credential',
            authentication: {
              api_key: {
                value: { plain_text: 'test-secret-value' }
              }
            }
          }
        ]
      }
    });
  });

  test('Complete CUJ: Create Config -> Instantiate -> Manage', async ({ page }) => {
    // 1. Navigate to Marketplace
    await page.goto('/marketplace');
    await expect(page.getByRole('heading', { name: 'Marketplace' })).toBeVisible();

    // 2. Open Wizard
    await page.getByRole('button', { name: 'Create Config' }).click();
    await expect(page.getByRole('dialog', { name: 'Create Upstream Service' })).toBeVisible();

    // 3. Step 1: Service Type
    await expect(page.getByText('Service Type')).toBeVisible();
    await page.getByRole('combobox').click();
    // Use exact name from backend seeds.go
    await page.getByRole('option', { name: 'PostgreSQL', exact: true }).click();
    await page.click('button:has-text("Next")');

    // 4. Step 2: Parameters
    await expect(page.getByRole('heading', { name: 'Environment Variables / Parameters' })).toBeVisible();

    // Check for parameter input existence and edit it
    // Using specific locator to avoid strict mode violations if multiple inputs exist
    // The value comes from the default in seeds.go
    const paramInput = page.locator('input[value="postgresql://postgres:postgres@localhost:5432/postgres"]');
    await expect(paramInput).toBeVisible();
    await paramInput.fill('postgresql://test:test@localhost:5432/testdb');

    // Add a new parameter
    await page.getByRole('button', { name: 'Add Parameter' }).click();

    // Wait for the new input to appear (should have 2 now: POSTGRES_URL + new one)
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

    // Verify we can see the credential we mocked (seeded)
    await page.getByRole('combobox').click();
    await expect(page.getByRole('option', { name: 'Test Credential' })).toBeVisible();
    // Select Test Credential
    await page.getByRole('option', { name: 'Test Credential' }).click();

    // Helper: Test Connection
    await page.getByRole('button', { name: 'Test Connection' }).click();
    // Expect success message
    await expect(page.getByText('Connection verification successful')).toBeVisible();

    await page.click('button:has-text("Next")');

    // 7. Step 5: Review
    await expect(page.getByText('Review & Finish')).toBeVisible();
    // Check if JSON contains our changes
    await expect(page.getByText('"MAX_CONNECTIONS"')).toBeVisible();
    await expect(page.getByText('"100"')).toBeVisible();
    await expect(page.getByText('postgresql://test:test@localhost:5432/testdb')).toBeVisible();

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

    await page.click('button:has-text("Create Instance")');

    // Verify toast or closing of dialog - wait for service registration
    await expect(page.getByRole('dialog', { name: 'Instantiate Service' })).toBeHidden({timeout: 10000});
  });
});
