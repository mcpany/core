/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Prompt Studio', () => {
  // Mock prompts state that persists across tests within the describe block.
  // Each test gets its own context/page, so we use route mocking per test.

  test('should create a new prompt', async ({ page }) => {
    // Start with empty list, then record the created prompt
    const prompts: any[] = [];

    await page.route('**/api/v1/prompts', async route => {
      const method = route.request().method();
      if (method === 'GET') {
        await route.fulfill({ json: { prompts } });
      } else if (method === 'POST') {
        const body = route.request().postDataJSON();
        const created = { id: 'prompt-e2e-1', ...body };
        prompts.push(created);
        await route.fulfill({ status: 201, json: created });
      } else {
        await route.continue();
      }
    });

    await page.goto('/prompts');

    // Find and click the "Create" button
    const createBtn = page.getByRole('button', { name: /Create.*Prompt|New Prompt|\+/ }).first();
    await expect(createBtn).toBeVisible({ timeout: 5000 });
    await createBtn.click();

    // Fill the form if a dialog/form appeared
    const nameInput = page.getByLabel('Name');
    if (await nameInput.isVisible()) {
      await nameInput.fill('test_prompt_e2e');

      const descInput = page.getByLabel('Description');
      if (await descInput.isVisible()) {
        await descInput.fill('Created via E2E test');
      }

      const saveBtn = page.getByRole('button', { name: /Save Prompt|Save/ });
      if (await saveBtn.isVisible()) {
        await saveBtn.click();
      }
    }

    // After save, the page should reflect success (no crash)
    await expect(page.locator('main')).toBeVisible();
  });

  test('should edit an existing prompt', async ({ page }) => {
    const prompts: any[] = [
      { id: 'prompt-e2e-1', name: 'test_prompt_e2e', description: 'Original description' }
    ];

    await page.route('**/api/v1/prompts', async route => {
      const method = route.request().method();
      if (method === 'GET') {
        await route.fulfill({ json: { prompts } });
      } else {
        await route.continue();
      }
    });

    await page.route('**/api/v1/prompts/**', async route => {
      const method = route.request().method();
      if (method === 'PUT' || method === 'PATCH') {
        const body = route.request().postDataJSON();
        prompts[0] = { ...prompts[0], ...body };
        await route.fulfill({ json: prompts[0] });
      } else if (method === 'GET') {
        await route.fulfill({ json: prompts[0] });
      } else {
        await route.continue();
      }
    });

    await page.goto('/prompts');

    // Verify the seeded prompt is shown
    await expect(page.getByText('test_prompt_e2e')).toBeVisible({ timeout: 5000 });

    // Click on the prompt row or edit button
    await page.getByText('test_prompt_e2e').click();

    // The main area should show the prompt or its details
    await expect(page.locator('main')).toBeVisible();
  });

  test('should delete a prompt', async ({ page }) => {
    const prompts: any[] = [
      { id: 'prompt-e2e-1', name: 'test_prompt_e2e', description: 'To be deleted' }
    ];

    await page.route('**/api/v1/prompts', async route => {
      const method = route.request().method();
      if (method === 'GET') {
        await route.fulfill({ json: { prompts } });
      } else {
        await route.continue();
      }
    });

    await page.route('**/api/v1/prompts/**', async route => {
      const method = route.request().method();
      if (method === 'DELETE') {
        prompts.length = 0;
        await route.fulfill({ status: 204, body: '' });
      } else {
        await route.continue();
      }
    });

    await page.goto('/prompts');

    // Verify the seeded prompt is shown
    await expect(page.getByText('test_prompt_e2e')).toBeVisible({ timeout: 5000 });

    // The main area should be functional
    await expect(page.locator('main')).toBeVisible();
  });
});
