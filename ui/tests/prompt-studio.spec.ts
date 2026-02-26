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

    // Seed a prompt for edit/delete tests
    const pResponse = await request.post('/api/v1/prompts', {
      data: {
        name: 'seed_prompt_e2e',
        description: 'Seeded for E2E tests',
        messages: [{
          role: 'USER',
          text: { text: 'Hello World' }
        }]
      }
    });
    if (!pResponse.ok()) {
      console.warn('Failed to seed prompt:', await pResponse.text());
    }
  });

  test.beforeEach(async ({ page }) => {
    // Navigate to Prompts page
    await page.goto('/prompts');
  });

  test('should create a new prompt', async ({ page }) => {
    page.on('request', request => {
      if (request.url().includes('/api/v1/')) {
        console.log('>>', request.method(), request.url());
        if (request.method() === 'POST' || request.method() === 'PUT') {
          console.log('>> BODY:', request.postData());
        }
      }
    });
    page.on('response', response => {
      if (response.url().includes('/api/v1/')) {
        console.log('<<', response.status(), response.url());
      }
    });
    page.on('console', msg => console.log('BROWSER LOG:', msg.text()));

    await page.goto('/prompts');
    // 1. Click "Create New Prompt" (or the + button in empty state)
    // We wait for the page to load and check if we are in empty state or list state
    // We look for any button that resembles "Create"
    const createBtn = page.getByRole('button', { name: /Create.*Prompt|New Prompt/ }).first();
    await createBtn.click();

    // 2. Fill the form
    // 2. Fill the form
    await page.getByLabel('Name').fill('test_prompt_create');
    await page.getByLabel('Description').fill('Created via E2E test');

    // Select Service
    // We expect 'E2E Test Service' to be in the list
    await page.getByRole('combobox', { name: 'Service' }).click();
    await page.getByRole('option', { name: 'E2E Test Service' }).click();

    // Fill Message
    await page.getByPlaceholder('Enter prompt text').fill('Hello {{name}}');

    // 3. Save
    console.log('Clicking Save Prompt...');
    await expect(page.getByRole('button', { name: 'Save Prompt' })).toBeEnabled();
    const savePromise = page.waitForResponse(resp =>
      resp.url().includes('/api/v1/services') &&
      resp.request().method() === 'PUT' &&
      resp.status() < 400
    );
    await page.getByRole('button', { name: 'Save Prompt' }).click();
    console.log('Clicked Save Prompt, waiting for response...');
    await savePromise;
    console.log('Save Prompt response received.');

    // 4. Verify it appears in the list
    await expect(async () => {
      // Reload the page to fetch the latest list
      await page.reload({ waitUntil: 'networkidle' });
      // Expect the prompt to be visible
      await expect(page.getByText('test_prompt_create')).toBeVisible();
      await expect(page.getByText('Created via E2E test')).toBeVisible();
    }).toPass({ timeout: 60000, intervals: [2000] });
  });

  test('should edit an existing prompt', async ({ page }) => {
    // Ensure the prompt exists (run sequential or seed prompt too)
    // For now we assume previous test ran or we re-create
    // But tests run in parallel by default? Use serial mode if needed or independent seeding.
    // Let's seed the prompt too in beforeAll or just rely on Create running first?
    // Playwright parallel execution is file-based usually. Tests in file run serial by default?
    // Default is serial within a file.

    // Wait for list to appear
    await expect(page.getByText('seed_prompt_e2e')).toBeVisible();
    await page.getByText('seed_prompt_e2e').click();

    // Click Edit button (Pencil icon)
    await page.locator('button').filter({ has: page.locator('svg.lucide-pencil') }).click();

    // Change Description
    await page.getByLabel('Description').fill('Updated description');

    // Save
    const savePromise = page.waitForResponse(resp =>
      resp.url().includes('/api/v1/services') &&
      resp.request().method() === 'PUT' &&
      resp.status() < 400
    );
    await page.getByRole('button', { name: 'Save Prompt' }).click();
    await savePromise;

    // Verify
    await expect(page.getByText('Updated description')).toBeVisible();
  });

  test('should delete a prompt', async ({ page }) => {
    // Select prompt
    await expect(page.getByText('seed_prompt_e2e')).toBeVisible();
    await page.getByText('seed_prompt_e2e').click();

    // Click Delete button (Trash icon)
    const deletePromise = page.waitForResponse(resp =>
      resp.url().includes('/api/v1/services') &&
      resp.request().method() === 'PUT' &&
      resp.status() < 400
    );
    await page.locator('button').filter({ has: page.locator('svg.lucide-trash-2') }).click();
    page.on('dialog', dialog => dialog.accept());
    await deletePromise;

    // Verify removal
    await expect(page.getByText('seed_prompt_e2e')).toBeHidden();
  });
});
