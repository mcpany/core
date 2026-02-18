/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('HTTP Tool Live Preview & Test', () => {
  const serviceName = 'live-preview-test-service';

  test.beforeEach(async ({ request }) => {
    // Cleanup first
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});

    // Seed HTTP service
    // We use httpbin.org for reliable echo testing
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

  test.afterEach(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should preview request and execute tool', async ({ page }) => {
    // 1. Navigate to Upstream Services
    await page.goto('/upstream-services');
    await expect(page.getByText(serviceName)).toBeVisible();

    // 2. Open Editor
    const row = page.getByRole('row', { name: serviceName });
    await row.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    // 3. Go to Tools -> Add Tool
    await page.getByRole('tab', { name: 'Tools' }).click();
    await page.getByRole('button', { name: 'Add Tool' }).click();

    // 4. Configure Tool
    await page.getByLabel('Tool Name').fill('get_uuid');
    await page.getByLabel('Description').fill('Get a UUID');
    await page.getByLabel('Endpoint Path').fill('/uuid');

    // 5. Open Live Test Tab
    await page.getByRole('tab', { name: 'Live Test' }).click();

    // 6. Verify Preview
    // Scope to the Request Preview area to avoid ambiguity
    const previewCard = page.locator('.rounded-lg', { hasText: 'Request Preview' });

    await expect(previewCard.getByText('GET', { exact: true })).toBeVisible();
    await expect(previewCard.getByText('https://httpbin.org/uuid')).toBeVisible();

    // 7. Execute (Pre-save)
    // Should be enabled because we passed serviceName, but it might fail or warn if backend needs registration first.
    // The UI button is enabled if serviceName is present.
    // But since the *tool* is not yet saved in the backend, the backend won't find `live-preview-test-service.get_uuid`.
    // So execution should fail with "Tool not found".
    await page.getByRole('button', { name: 'Execute' }).click();
    // We expect an error message (red text) inside the result box
    // Use .last() or specific container to avoid other alerts
    await expect(page.locator('.bg-muted\\/30 .text-destructive')).toBeVisible();

    // 8. Save Service (and Tool)
    // Close Tool Editor
    await page.keyboard.press('Escape');
    // Save Service
    await page.getByRole('button', { name: 'Save Changes' }).click();
    await expect(page.getByText('Service Updated')).toBeVisible();

    // 9. Re-open Tool Editor
    // Wait for list to refresh or check if we are still in sheet?
    // Saving usually closes the sheet.
    // Re-open edit
    const row2 = page.getByRole('row', { name: serviceName });
    await row2.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();
    await page.getByRole('tab', { name: 'Tools' }).click();
    await page.getByRole('button', { name: 'Edit' }).click();

    // 10. Execute (Post-save)
    await page.getByRole('tab', { name: 'Live Test' }).click();
    await page.getByRole('button', { name: 'Execute' }).click();

    // 11. Verify Result
    // httpbin /uuid returns { "uuid": "..." }
    await expect(page.getByText('"uuid":')).toBeVisible();
  });

  test('should map path parameters in preview', async ({ page }) => {
    // 1. Navigate and Open Editor (Reuse existing service)
    await page.goto('/upstream-services');
    const row = page.getByRole('row', { name: serviceName });
    await row.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();
    await page.getByRole('tab', { name: 'Tools' }).click();
    await page.getByRole('button', { name: 'Add Tool' }).click();

    // 2. Configure with Path Param
    await page.getByLabel('Tool Name').fill('get_user');
    await page.getByLabel('Endpoint Path').fill('/users/{id}');

    // 3. Add Parameter definition
    await page.getByRole('button', { name: 'Add Parameter' }).click();
    await page.getByLabel('Name', { exact: true }).fill('id');

    // 4. Live Test
    await page.getByRole('tab', { name: 'Live Test' }).click();

    // Initial state: /users/{id}
    const previewCard = page.locator('.rounded-lg', { hasText: 'Request Preview' });
    // Scope search to preview card to distinguish from input field
    await expect(previewCard.getByText('/users/{id}')).toBeVisible();

    // 5. Type Arguments
    await page.locator('textarea').fill('{"id": 123}');

    // 6. Verify Preview Update
    await expect(previewCard.getByText('https://httpbin.org/users/123')).toBeVisible();
  });
});
