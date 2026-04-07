/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Prompt Studio Bulk Actions', () => {
  const serviceName = 'E2E Test Bulk Service';

  test.beforeAll(async ({ request }) => {
    // Delete if exists
    await request.delete(`/api/v1/services/e2e-test-bulk-service`).catch(() => {});

    // Seed a service for the test to ensure we have a place to save prompts
    const response = await request.post('/api/v1/services', {
      data: {
        id: 'e2e-test-bulk-service',
        name: serviceName,
        command_line_service: {
          command: 'echo',
          working_directory: '/tmp',
          prompts: [
            {
              name: 'test_prompt_bulk_1',
              title: 'Test Prompt 1',
              description: 'Test Bulk 1',
              disable: false
            },
            {
              name: 'test_prompt_bulk_2',
              title: 'Test Prompt 2',
              description: 'Test Bulk 2',
              disable: false
            }
          ]
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
    // Wait for the seeded prompts to appear in the list
    await expect(page.getByText('test_prompt_bulk_1').first()).toBeVisible();
    await expect(page.getByText('test_prompt_bulk_2').first()).toBeVisible();

    // Select both prompts using their respective checkboxes
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
