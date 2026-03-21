/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Prompt Studio', () => {
  test.beforeAll(async ({ request }) => {
    const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
    const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };

    // Seed a service for the test
    const response = await request.post('/api/v1/services', {
      data: {
        id: 'e2e-test-service',
        name: 'E2E Test Service',
        command_line_service: {
          command: 'echo',
          working_directory: '/tmp'
        },
        disable: false
      },
      headers: HEADERS
    });
    // We expect 201 Created or 200 OK. Even 400 if it already exists is fine-ish?
    // Better to ensure it works.
    if (!response.ok()) {
      console.warn('Failed to seed service:', await response.text());
      // Attempt to proceed anyway, maybe it exists
    }
  });

  test.beforeEach(async ({ page, request }) => {
    const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
    const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };

    const user = {
        id: 'e2e-admin-core',
        authentication: {
            basic_auth: {
                username: 'e2e-admin-core',
                password_hash: "$2a$12$KPRtQETm7XKJP/L6FjYYxuCFpTK/oRs7v9U6hWx9XFnWy6UuDqK/a" // password
            }
        },
        roles: ["admin"],
        profile_ids: ["dev"]
    };

    await request.post('/api/v1/users', { data: user, headers: HEADERS });

    // Login
    await page.goto('/login');
    await page.fill('input[name="username"]', 'e2e-admin-core');
    await page.fill('input[name="password"]', 'password');
    await Promise.all([
      page.waitForURL('/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
    ]);

    // Navigate to Prompts page
    await page.goto('/prompts');
  });

  test('should create a new prompt', async ({ page }) => {
    // 1. Click "Create New Prompt" (or the + button in empty state)
    // We wait for the page to load and check if we are in empty state or list state
    // We look for any button that resembles "Create"
    const createBtn = page.getByRole('button', { name: /Create.*Prompt/ }).first();
    await createBtn.click();

    // 2. Fill the form
    await page.getByLabel('Name').fill('test_prompt_e2e');
    await page.getByLabel('Description').fill('Created via E2E test');

    // Select Service
    // We expect 'E2E Test Service' to be in the list
    await page.getByRole('combobox', { name: 'Service' }).click();
    await page.getByRole('option', { name: 'E2E Test Service' }).click();

    // Fill Message
    await page.getByPlaceholder('Enter prompt text').fill('Hello {{name}}');

    const savePromise = page.waitForResponse(response =>
      response.url().includes('/api/v1/services/') &&
      response.request().method() === 'PUT' &&
      response.status() === 200,
    );

    // 3. Save
    await page.getByRole('button', { name: 'Save Prompt' }).click();
    await savePromise;

    // 4. Verify we return to prompt library successfully
    await expect(page).toHaveURL(/\/prompts\/?$/);
  });
});
