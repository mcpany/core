/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('HTTP Tool Live Preview', () => {
  const serviceName = 'preview-test-service';

  test.beforeEach(async ({ request }) => {
    // Cleanup
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});

    // Seed
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        http_service: {
            address: "https://api.example.com"
        }
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterEach(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should live preview parameter substitution', async ({ page }) => {
    await page.goto('/upstream-services');

    // Find row, click edit
    const row = page.getByRole('row', { name: serviceName });
    await row.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    // Go to Tools
    await page.getByRole('tab', { name: 'Tools' }).click();
    await page.getByRole('button', { name: 'Add Tool' }).click();

    // Setup Tool
    await page.getByLabel('Tool Name').fill('get_user');
    await page.getByLabel('Endpoint Path').fill('/users/{id}');

    // Add Parameter
    await page.getByRole('button', { name: 'Add Parameter' }).click();
    await page.getByLabel('Name', { exact: true }).fill('id');

    // Test Arguments
    await page.getByPlaceholder('{}').fill('{\n  "id": "123"\n}');

    // Verify Preview
    // We expect the URL in the preview card to be updated
    // The Preview displays "GET" and the URL.
    await expect(page.getByText('/users/123')).toBeVisible();
    await expect(page.getByText('https://api.example.com/users/123')).toBeVisible();
  });
});
