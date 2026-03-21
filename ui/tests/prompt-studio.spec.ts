/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Prompt Studio', () => {
  test.beforeAll(async ({ request }) => {
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
      }
    });
    // We expect 201 Created or 200 OK. Even 400 if it already exists is fine-ish?
    // Better to ensure it works.
    if (!response.ok()) {
      console.warn('Failed to seed service:', await response.text());
      // Attempt to proceed anyway, maybe it exists
    }
  });

  test.beforeEach(async ({ page }) => {
    // Navigate to Prompts page
    await page.goto('/prompts');
  });

  test('should create a new prompt', async ({ page }) => {
    // 1. Click "Create New Prompt" (or the + button in empty state)
    // We wait for the page to load and check if we are in empty state or list state
    // We look for any button that resembles "Create"
    const createBtn = page.getByRole('button', { name: /Create.*Prompt|New Prompt/ }).first();
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

  test('should verify prompt execution with RichResultViewer', async ({ page, request }) => {
    // Seed a specific prompt for execution test
    const res = await request.post('/api/v1/prompts', {
      data: {
        name: "rich_test_prompt",
        description: "A prompt for testing RichResultViewer",
        service_id: "e2e-test-service",
        disable: false,
        input_schema: "{\n  \"type\": \"object\",\n  \"properties\": {\n    \"arg1\": { \"type\": \"string\" }\n  }\n}",
        messages: [
            { role: "user", content: "Test content {{arg1}}" }
        ]
      }
    });

    if (!res.ok() && res.status() !== 409) {
        console.warn('Failed to seed rich_test_prompt:', await res.text());
    }

    await page.goto('/prompts');

    // Select the prompt
    // Wait for the prompt list to render
    await expect(page.getByText('rich_test_prompt').first()).toBeVisible({ timeout: 10000 });
    await page.getByText('rich_test_prompt').first().click();

    // Fill the argument
    await page.getByLabel('arg1').fill('RichResultTest');

    // Intercept the execution request and mock a rich response
    await page.route('**/api/v1/prompts/rich_test_prompt/execute', async route => {
      const json = {
        messages: [
          {
            role: "assistant",
            content: {
              type: "text",
              text: "This is a **rich** response."
            }
          }
        ]
      };
      await route.fulfill({ json });
    });

    // Execute
    await page.getByRole('button', { name: 'Generate Preview' }).click();

    // Verify RichResultViewer is used (it renders Markdown)
    await expect(page.locator('.prose')).toBeVisible();
    await expect(page.locator('strong:has-text("rich")')).toBeVisible();
  });
});
