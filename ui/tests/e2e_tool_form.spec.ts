/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('Tool Inspector Form Generation', async ({ page }) => {
  // Mock the services list API (dependency for tools page)
  await page.route('**/api/v1/services*', async route => {
      await route.fulfill({ json: [] });
  });

  // Mock the tool usage API
  await page.route('**/api/v1/tools/usage*', async route => {
      await route.fulfill({ json: [] });
  });

  // Mock the tools list API
  await page.route('**/api/v1/tools*', async route => {
    const json = {
      tools: [{
        name: 'test-tool',
        description: 'A test tool',
        serviceId: 'test-service',
        inputSchema: {
          type: 'object',
          properties: {
            message: { type: 'string', description: 'The message to echo' },
            count: { type: 'number', default: 1 },
            isUrgent: { type: 'boolean', description: 'Is this urgent?' }
          },
          required: ['message']
        }
      }]
    };
    await route.fulfill({ json });
  });

  // Mock the execution API
  await page.route('**/api/v1/tools/execute', async route => {
      const postData = route.request().postDataJSON();
      // Verify arguments were passed correctly
      if (postData.toolName === 'test-tool' &&
          postData.arguments.message === 'Hello World' &&
          postData.arguments.count === 5 &&
          postData.arguments.isUrgent === true) {
          await route.fulfill({ json: { result: "Success" } });
      } else {
          await route.fulfill({ status: 400, json: { error: "Invalid arguments" } });
      }
  });

  // Mock audit logs
  await page.route('**/api/v1/audit/logs*', async route => {
       await route.fulfill({ json: { entries: [] } });
  });

  // Navigate to tools page
  // We use networkidle to ensure initial data loads
  await page.goto('/tools', { waitUntil: 'networkidle' });

  // Click on the tool card to open inspector
  // The tool name should be visible
  await page.getByText('test-tool').click();

  // Switch to Form tab (it might be default, but let's click to be sure)
  // Note: The Inspector opens in a Dialog.
  const formTab = page.getByRole('tab', { name: 'Form' });
  await expect(formTab).toBeVisible();
  await formTab.click();

  // Verify form fields
  await expect(page.getByLabel('message')).toBeVisible();
  await expect(page.getByLabel('count')).toBeVisible();
  await expect(page.getByLabel('isUrgent')).toBeVisible();

  // Fill form
  await page.getByLabel('message').fill('Hello World');
  await page.getByLabel('count').fill('5');
  // Switch click toggles it
  await page.getByLabel('isUrgent').click();

  // Verify JSON updates (switch to JSON tab)
  await page.getByRole('tab', { name: 'JSON' }).click();
  const jsonContent = await page.locator('textarea').inputValue();
  const parsed = JSON.parse(jsonContent);
  expect(parsed.message).toBe('Hello World');
  expect(parsed.count).toBe(5);
  expect(parsed.isUrgent).toBe(true);

  // Switch back to Form and Execute
  await page.getByRole('tab', { name: 'Form' }).click();

  // Click Execute
  await page.getByRole('button', { name: 'Execute' }).click();

  // Verify Result
  await expect(page.getByText('"result": "Success"')).toBeVisible();
});
