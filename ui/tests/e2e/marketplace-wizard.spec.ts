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
    const templates: any[] = [
        {
            id: 'postgres',
            name: 'PostgreSQL',
            version: '1.0.0',
            command_line_service: {
                command: 'npx -y @modelcontextprotocol/server-postgres',
                env: {
                    "POSTGRES_URL": { plain_text: "postgresql://user:password@localhost:5432/dbname" }
                }
            },
            configuration_schema: JSON.stringify({
                type: "object",
                properties: {
                    "POSTGRES_URL": { type: "string", title: "Connection URL" }
                }
            })
        }
    ];
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
    await page.getByRole('button', { name: 'Create Config' }).click();
    await expect(page.getByRole('dialog', { name: 'Create Upstream Service' })).toBeVisible();

    // 3. Step 1: Service Type
    await expect(page.getByText('Service Type')).toBeVisible();
    await page.getByRole('combobox').click();
    await page.getByRole('option', { name: 'PostgreSQL' }).click();

    // Verify selection took effect (Service Name should auto-fill)
    await expect(page.locator('#service-name')).toHaveValue('PostgreSQL');

    await page.click('button:has-text("Next")');

    // Verify transition to Step 2
    await expect(page.getByText('2. Configure Parameters')).toBeVisible();

    // 4. Step 2: Parameters
    await expect(page.getByRole('heading', { name: 'Configuration Parameters' })).toBeVisible();

    // Check for parameter input existence and edit it
    // Schema form uses labels from schema
    const paramInput = page.getByLabel('Connection URL');
    await expect(paramInput).toBeVisible();
    await paramInput.fill('postgresql://test:test@localhost:5432/testdb');

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
