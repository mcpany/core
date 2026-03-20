/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Marketplace Wizard and Service Lifecycle', () => {

  const testCredId = `cred-${Date.now()}`;
  const templateId = `tpl-${Date.now()}`;

  test.beforeAll(async ({ request }) => {
    // Seed real credential
    const credResponse = await request.post('/api/v1/credentials', {
      data: {
        id: testCredId,
        name: 'Test Credential',
        authentication: { apiKey: { paramName: 'Authorization', in: 0, value: { plainText: 'secret' } } }
      }
    });
    // Ignore error if already exists, just checking seeding ok
    if (!credResponse.ok()) {
       console.log('Credential seed failed, might exist or backend error', await credResponse.text());
    }

    // Seed Template
    const templateResponse = await request.post('/api/v1/templates', {
      data: {
        id: templateId,
        name: 'PostgreSQL Database E2E Wizard',
        description: 'Read-only access to PostgreSQL databases',
        service_config: {
          name: 'PostgreSQL Database',
          command_line_service: {
            command: 'npx -y @modelcontextprotocol/server-postgres',
            env: {
              POSTGRES_URL: {
                plain_text: 'postgresql://user:password@localhost:5432/dbname',
              },
            },
            working_directory: '',
            tools: [],
            resources: [],
            prompts: [],
            calls: {},
            communication_protocol: 0,
            local: false,
          },
        },
        params: {
          POSTGRES_URL: 'postgresql://user:password@localhost:5432/dbname',
        },
      }
    });
    if (!templateResponse.ok()) {
       console.log('Template seed failed', await templateResponse.text());
    }
  });

  test.afterAll(async ({ request }) => {
    await request.delete(`/api/v1/credentials/${testCredId}`).catch(() => {});
    await request.delete(`/api/v1/templates/${templateId}`).catch(() => {});
  });

  test('Complete CUJ: Create Config -> Instantiate -> Manage', async ({ page }) => {
    // Note: Do not mock /debug/auth-test or other routes if we can avoid it.
    // Auth test might fail if the real credential/backend logic fails.
    // Let's mock just the auth-test because it connects to real APIs externally usually,
    // or we'll accept whatever status it returns if we can't seed external servers.
    await page.route('/api/v1/debug/auth-test', async route => {
      await route.fulfill({ json: { success: true, message: "Connection verification successful" } });
    });

    // 1. Navigate to Marketplace
    await page.goto('/marketplace');
    await expect(page.getByText('Marketplace', { exact: true }).first()).toBeVisible();

    // 2. Open Wizard
    await page.getByRole('button', { name: 'Create Config' }).click();
    await expect(page.getByRole('dialog', { name: 'Create Upstream Service' })).toBeVisible();

    // 3. Step 1: Service Type
    await expect(page.getByText('Service Type')).toBeVisible();
    await page.getByRole('combobox').click();
    await page.getByRole('option', { name: 'PostgreSQL Database E2E Wizard' }).click();
    await page.click('button:has-text("Next")');

    // 4. Step 2: Parameters
    await expect(page.getByPlaceholder('VAR_NAME')).toBeVisible();

    // Check for parameter input existence and edit it
    // Use the Value placeholder input in the first row
    const paramInput = page.locator('input[placeholder="Value"]').first();
    await expect(paramInput).toHaveValue('postgresql://user:password@localhost:5432/dbname');
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

    // Verify we can see the credential we mocked
    await page.getByRole('combobox').click({ force: true });
    await expect(page.getByRole('option', { name: 'Test Credential' })).toBeVisible({ timeout: 10000 });
    // Select Test Credential
    await page.getByRole('option', { name: 'Test Credential' }).click();

    // Helper: Test Connection
    await page.getByRole('button', { name: 'Test Connection' }).click();
    // Expect success message (toast or alert or status)
    await expect(page.getByText('Connection verification successful')).toBeVisible({ timeout: 60000 });

    await page.click('button:has-text("Next")');

    // 7. Step 5: Review
    await expect(page.getByText('Review & Finish')).toBeVisible(); // Title is "5. Review & Finish" in create-config-wizard.tsx
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
