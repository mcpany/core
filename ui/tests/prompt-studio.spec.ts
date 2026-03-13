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
        commandLineService: {
            command: 'echo',
            workingDirectory: '/tmp'
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

    // 3. Save
    await page.getByRole('button', { name: 'Save Prompt' }).click();

    // 4. Verify it appears in the list
    await expect(page.getByText('test_prompt_e2e')).toBeVisible();
    await expect(page.getByText('Created via E2E test')).toBeVisible();
  });

  test('should edit an existing prompt', async ({ page }) => {
    // Ensure the prompt exists (run sequential or seed prompt too)
    // For now we assume previous test ran or we re-create
    // But tests run in parallel by default? Use serial mode if needed or independent seeding.
    // Let's seed the prompt too in beforeAll or just rely on Create running first?
    // Playwright parallel execution is file-based usually. Tests in file run serial by default?
    // Default is serial within a file.

    // Wait for list to appear
    await expect(page.getByText('test_prompt_e2e')).toBeVisible();
    await page.getByText('test_prompt_e2e').click();

    // Click Edit button (Pencil icon)
    await page.locator('button').filter({ has: page.locator('svg.lucide-pencil') }).click();

    // Change Description
    await page.getByLabel('Description').fill('Updated description');

    // Save
    await page.getByRole('button', { name: 'Save Prompt' }).click();

    // Verify
    await expect(page.getByText('Updated description')).toBeVisible();
  });

  test('should delete a prompt', async ({ page }) => {
    // Select prompt
    await expect(page.getByText('test_prompt_e2e')).toBeVisible();
    await page.getByText('test_prompt_e2e').click();

    // Click Delete button (Trash icon)
    await page.locator('button').filter({ has: page.locator('svg.lucide-trash-2') }).click();

    // Confirm dialog
    page.on('dialog', dialog => dialog.accept());

    // Verify removal
    await expect(page.getByText('test_prompt_e2e')).toBeHidden();
  });
});
