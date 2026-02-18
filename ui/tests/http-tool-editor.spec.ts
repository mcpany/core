/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('HTTP Tool Editor - Live Preview', () => {
  const serviceName = 'e2e-preview-test';

  test.beforeAll(async ({ request }) => {
    // Seed HTTP service pointing to httpbin
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});

    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        http_service: {
            address: "https://httpbin.org"
        }
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should preview request and execute tool', async ({ page, request }) => {
    // Navigate
    await page.goto('/upstream-services');
    await expect(page.getByText(serviceName)).toBeVisible();

    const row = page.getByRole('row', { name: serviceName });
    await row.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    // Open Tools
    await page.getByRole('tab', { name: 'Tools' }).click();
    await page.getByRole('button', { name: 'Add Tool' }).click();

    // Configure Tool
    await page.getByLabel('Tool Name').fill('get_uuid');
    await page.getByLabel('Endpoint Path').fill('/uuid');

    // Check Preview (Immediate)
    await page.getByRole('tab', { name: 'Test & Preview' }).click();

    // Verify Preview Content
    // Should show GET and /uuid
    await expect(page.locator('.text-blue-500').getByText('GET')).toBeVisible();
    // Note: The badge classes might vary, but text content is reliable
    await expect(page.getByText('/uuid')).toBeVisible();

    // Add a parameter to verify substitution
    await page.getByRole('tab', { name: 'Request Parameters' }).click();
    await page.getByLabel('Endpoint Path').fill('/uuid/{id}');
    await page.getByRole('button', { name: 'Add Parameter' }).click();
    await page.getByLabel('Name', { exact: true }).fill('id');

    // Go back to Preview
    await page.getByRole('tab', { name: 'Test & Preview' }).click();

    // Verify Substitution Placeholder
    await expect(page.getByText('/uuid/{id}')).toBeVisible();

    // Type Argument
    const argsInput = page.getByLabel('Test Arguments (JSON)');
    await argsInput.fill('{\n  "id": "123"\n}');

    // Verify Substitution Result
    await expect(page.getByText('/uuid/123')).toBeVisible();

    // Now Save and Test Execution
    // Close Tool Editor
    await page.keyboard.press('Escape');

    // Save Service
    await page.getByRole('button', { name: 'Save Changes' }).click();
    await expect(page.getByText('Service Updated')).toBeVisible();

    // Re-open Tool
    await page.getByRole('button', { name: 'Edit' }).first().click(); // Assuming it's the only/first tool
    await page.getByRole('tab', { name: 'Test & Preview' }).click();

    // Execute
    // Fill args again as state might reset (or I could persist it, but for now simple)
    // Actually, state resets on re-mount.
    // Wait, get_uuid on httpbin returns a UUID. The /uuid/{id} endpoint doesn't exist on httpbin.
    // httpbin has /uuid.
    // Let's use /anything/{id} which returns args.

    // Fix: Update path to /anything/{id}
    await page.getByRole('tab', { name: 'Request Parameters' }).click();
    await page.getByLabel('Endpoint Path').fill('/anything/{id}');

    // Close & Save again (to update definition)
    await page.keyboard.press('Escape');
    await page.getByRole('button', { name: 'Save Changes' }).click();
    await expect(page.getByText('Service Updated')).toBeVisible(); // Wait for toast

    // Re-open
    await page.getByRole('button', { name: 'Edit' }).first().click();
    await page.getByRole('tab', { name: 'Test & Preview' }).click();

    // Fill Args
    await page.getByLabel('Test Arguments (JSON)').fill('{\n  "id": "test-execution"\n}');

    // Click Execute
    await page.getByRole('button', { name: 'Execute' }).click();

    // Verify Loading
    // await expect(page.getByText('Executing...')).toBeVisible(); // Might be too fast

    // Verify Result
    // httpbin /anything/{id} returns json with url including path
    await expect(page.getByText('test-execution')).toBeVisible();
    await expect(page.getByText('https://httpbin.org/anything/test-execution')).toBeVisible();
  });
});
