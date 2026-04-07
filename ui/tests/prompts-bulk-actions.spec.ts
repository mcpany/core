/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Prompt Studio Bulk Actions', () => {
  const serviceName = 'E2E Test Bulk Service';

  test.beforeAll(async ({ request }) => {
    // Seed a service for the test to ensure we have a place to save prompts
    const response = await request.post('/api/v1/services', {
      data: {
        id: 'e2e-test-bulk-service',
        name: serviceName,
        command_line_service: {
          command: 'echo',
          working_directory: '/tmp'
        },
        disable: false
      }
    });
    if (!response.ok()) {
      console.warn('Failed to seed service:', await response.text());
    }
  });

  test.afterAll(async ({ request }) => {
    // Clean up
    await request.delete(`/api/v1/services/e2e-test-bulk-service`);
  });

  test.beforeEach(async ({ page }) => {
    // Navigate to Prompts page
    await page.goto('/prompts');
  });

  test('should support bulk enabling and disabling of prompts', async ({ page }) => {
    // 1. Create first prompt
    let createBtn = page.getByRole('button', { name: /Create.*Prompt|New Prompt/ }).first();
    await createBtn.click();
    await page.getByLabel('Name').fill('test_prompt_bulk_1');
    await page.getByLabel('Description').fill('Test Bulk 1');
    await page.getByRole('combobox', { name: 'Service' }).click();
    await page.getByRole('option', { name: serviceName }).click();
    await page.getByPlaceholder('Enter prompt text').fill('Hello {{name}}');

    let savePromise = page.waitForResponse(response =>
      response.url().includes('/api/v1/services/') &&
      response.request().method() === 'PUT' &&
      response.status() === 200,
    );
    await page.getByRole('button', { name: 'Save Prompt' }).click();
    await savePromise;

    // 2. Create second prompt
    createBtn = page.getByRole('button', { name: /Create.*Prompt|New Prompt/ }).first();
    await createBtn.click();
    await page.getByLabel('Name').fill('test_prompt_bulk_2');
    await page.getByLabel('Description').fill('Test Bulk 2');
    await page.getByRole('combobox', { name: 'Service' }).click();
    await page.getByRole('option', { name: serviceName }).click();
    await page.getByPlaceholder('Enter prompt text').fill('Hello {{name}}');

    savePromise = page.waitForResponse(response =>
      response.url().includes('/api/v1/services/') &&
      response.request().method() === 'PUT' &&
      response.status() === 200,
    );
    await page.getByRole('button', { name: 'Save Prompt' }).click();
    await savePromise;

    // Wait for the prompts to appear in the list
    await expect(page.getByText('test_prompt_bulk_1').first()).toBeVisible();
    await expect(page.getByText('test_prompt_bulk_2').first()).toBeVisible();

    // 3. Select both prompts using their respective checkboxes
    // The checkbox is adjacent to the button containing the prompt name.
    // We can locate the checkbox by looking at the parent group.
    await page.locator('.group').filter({ hasText: 'test_prompt_bulk_1' }).getByRole('checkbox').click();
    await page.locator('.group').filter({ hasText: 'test_prompt_bulk_2' }).getByRole('checkbox').click();

    // 4. Verify bulk actions appear
    await expect(page.getByText('2 selected')).toBeVisible();

    // 5. Bulk Disable
    const disableBtn = page.getByRole('button', { name: 'Disable', exact: true });
    await expect(disableBtn).toBeVisible();

    // Set up response waiting for disable calls. There should be two calls.
    let responseCount = 0;
    const disablePromise = page.waitForResponse(res => {
      if (res.url().includes('/api/v1/prompts') && res.request().method() === 'POST' && res.status() === 200) {
        responseCount++;
        return responseCount === 2;
      }
      return false;
    });

    await disableBtn.click();
    await disablePromise;

    // 6. Verify visual update (line-through is applied when disabled)
    // Both prompt names should now have the line-through class
    await expect(page.locator('.group').filter({ hasText: 'test_prompt_bulk_1' }).locator('span.line-through')).toBeVisible();
    await expect(page.locator('.group').filter({ hasText: 'test_prompt_bulk_2' }).locator('span.line-through')).toBeVisible();

    // 7. Bulk Enable
    const enableBtn = page.getByRole('button', { name: 'Enable', exact: true });
    await expect(enableBtn).toBeVisible();

    let enableCount = 0;
    const enablePromise = page.waitForResponse(res => {
      if (res.url().includes('/api/v1/prompts') && res.request().method() === 'POST' && res.status() === 200) {
        enableCount++;
        return enableCount === 2;
      }
      return false;
    });

    await enableBtn.click();
    await enablePromise;

    // 8. Verify visual update (line-through is removed)
    await expect(page.locator('.group').filter({ hasText: 'test_prompt_bulk_1' }).locator('span.line-through')).not.toBeVisible();
    await expect(page.locator('.group').filter({ hasText: 'test_prompt_bulk_2' }).locator('span.line-through')).not.toBeVisible();
  });
});
