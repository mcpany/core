/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Marketplace Wizard and Service Lifecycle', () => {

  test.beforeEach(async ({ page }) => {
    // Mock API responses
    await page.route('/api/v1/services', async route => {
      if (route.request().method() === 'GET') {
          await route.fulfill({ json: [] });
      } else if (route.request().method() === 'POST') {
          await route.fulfill({ json: { status: 'success' } });
      } else {
        await route.continue();
      }
    });

    await page.route('/api/v1/marketplace/official', async route => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify([]) });
    });

    await page.route('/api/v1/marketplace/public', async route => {
        await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify([]) });
    });

    await page.route('/api/v1/credentials', async route => {
        await route.fulfill({ json: [{ id: 'cred-1', name: 'Test Credential' }] });
    });

    // Mock Templates API
    const templates: any[] = [];
    await page.route('/api/v1/templates', async route => {
        if (route.request().method() === 'GET') {
            await route.fulfill({ json: templates });
        } else if (route.request().method() === 'POST') {
             const data = await route.request().postDataJSON();
             templates.push({ ...data, id: `tpl-${Date.now()}` });
             await route.fulfill({ json: {} });
        } else {
            await route.continue();
        }
    });

    await page.route('/api/v1/templates/*', async route => {
        if (route.request().method() === 'DELETE') {
             // Basic mock
             await route.fulfill({ json: {} });
        } else {
             await route.continue();
        }
    });

    // Mock Auth Test
    await page.route('/api/v1/debug/auth-test', async route => {
        await route.fulfill({ json: { success: true, message: "Connection verification successful" } });
    });
  });

  test('Complete CUJ: Create Config -> Instantiate -> Manage', async ({ page }) => {
    // 1. Navigate to Marketplace
    await page.goto('/marketplace');
    await expect(page.getByRole('heading', { name: 'Marketplace' })).toBeVisible();

    // 2. Open Wizard
    // Use force click to ensure it clicks even if something is overlaying it slightly (though unexpected)
    await page.getByRole('button', { name: 'Create Config' }).click({ force: true });

    // Allow more time for dialog animation/rendering
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Create Upstream Service Config')).toBeVisible();

    // 3. Step 1: Service Type
    await expect(page.getByText('Service Type')).toBeVisible();
    await page.getByRole('combobox').click();
    await page.getByRole('option', { name: 'PostgreSQL' }).click();
    await page.click('button:has-text("Next")');

    // 4. Step 2: Parameters
    // With schema, the heading is "Configuration"
    await expect(page.getByRole('heading', { name: 'Configuration' })).toBeVisible();

    // Check for parameter input existence and edit it
    // Schema form uses labels based on schema title "Connection URL"
    const paramInput = page.getByLabel('Connection URL');
    await expect(paramInput).toBeVisible();
    await paramInput.fill('postgresql://test:test@localhost:5432/testdb');

    // Schema-based forms don't have "Add Parameter" button by default unless schema allows additional properties (which SchemaForm currently doesn't implement)
    // The PostgreSQL schema requires POSTGRES_URL.

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

    // Verify we can see the credential we mocked
    await page.getByRole('combobox').click();
    await expect(page.getByRole('option', { name: 'Test Credential' })).toBeVisible();
    // Select Test Credential
    await page.getByRole('option', { name: 'Test Credential' }).click();

    // Helper: Test Connection
    await page.getByRole('button', { name: 'Test Connection' }).click();
    // Expect success message (toast or alert or status)
    await expect(page.getByText('Connection verification successful')).toBeVisible();

    await page.click('button:has-text("Next")');

    // 7. Step 5: Review
    await expect(page.getByText('Review & Finish')).toBeVisible(); // Title is "5. Review & Finish" in create-config-wizard.tsx
    // Check if JSON contains our changes
    // With SchemaForm, POSTGRES_URL is updated
    await expect(page.getByText('postgresql://test:test@localhost:5432/testdb')).toBeVisible();

    await page.click('button:has-text("Finish & Save")');

    // 8. Verify Saved to Local
    // Wait for dialog to close
    await expect(page.getByRole('dialog')).toBeHidden();

    await expect(page.getByRole('tab', { name: 'Local' })).toBeVisible();
    await page.getByRole('tab', { name: 'Local' }).click();

    // 9. Instantiate
    // There might be multiple "Instantiate" buttons if multiple templates exist.
    // Assuming our new template is the first/only one in mock.
    await expect(page.getByRole('button', { name: 'Instantiate' }).first()).toBeVisible();
    await page.getByRole('button', { name: 'Instantiate' }).first().click();

    await expect(page.getByRole('dialog', { name: 'Instantiate Service' })).toBeVisible();

    // The template name might be "My Postgres DB" (default) or whatever I typed if I did.
    // In Step 1, I didn't change name, so it's likely "PostgreSQL Database" or whatever StepServiceType defaults to.

    const nameInput = page.locator('#service-name-input');
    await expect(nameInput).toBeVisible();

    const uniqueName = `postgres-test-${Date.now()}`;
    await nameInput.fill(uniqueName);

    // Schema form in instantiate dialog
    const instantiateParamInput = page.getByLabel('Connection URL');
    await expect(instantiateParamInput).toBeVisible();
    // It should have the value from template
    await expect(instantiateParamInput).toHaveValue('postgresql://test:test@localhost:5432/testdb');

    // Mock the register service call
    // Wait for the response after clicking
    const registerPromise = page.waitForResponse(response =>
        response.url().includes('/api/v1/services') && response.status() === 200
    );

    await page.click('button:has-text("Create Instance")');
    await registerPromise;

    // Verify toast or closing of dialog
    await expect(page.getByRole('dialog', { name: 'Instantiate Service' })).toBeHidden();
  });
});
