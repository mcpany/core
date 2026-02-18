/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('HTTP Tool Live Test', () => {
  const serviceName = 'e2e-live-test-service';

  test.beforeAll(async ({ request }) => {
    // Seed HTTP service pointing to httpbin for real execution testing
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

  test('should preview and execute HTTP tool', async ({ page, request }) => {
    // Navigate to upstream services
    await page.goto('/upstream-services');

    // Wait for list to load
    await expect(page.getByText(serviceName)).toBeVisible();

    // Find row, click edit.
    const row = page.getByRole('row', { name: serviceName });
    await row.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    // Click "Tools" tab.
    await page.getByRole('tab', { name: 'Tools' }).click();

    // Add Tool
    await page.getByRole('button', { name: 'Add Tool' }).click();

    // Configure Tool
    await page.getByLabel('Tool Name').fill('get_uuid');
    await page.getByLabel('Endpoint Path').fill('/uuid'); // httpbin.org/uuid returns a UUID

    // Add dummy parameter to test preview substitution
    await page.getByRole('button', { name: 'Add Parameter' }).click();
    await page.getByLabel('Name', { exact: true }).fill('foo');

    // We want to test query param mapping. By default, it maps to query if not in path.

    // ---------------------
    // Test Preview
    // ---------------------

    // Type in Test Arguments
    // Note: The textarea placeholders might overlap, so we click and type.
    await page.getByLabel('Test Arguments (JSON)').fill('{"foo": "bar"}');

    // Verify Preview
    // Should show GET https://httpbin.org/uuid?foo=bar
    await expect(page.getByText('GET')).toBeVisible();
    await expect(page.getByText('https://httpbin.org/uuid?foo=bar')).toBeVisible();

    // ---------------------
    // Test Execution
    // ---------------------

    // Execution requires saving first.
    // We check the warning message.
    await expect(page.getByText('Note: You must Save Changes')).toBeVisible();

    // We can't save from within the Tool Editor sheet easily without closing it in the current UI flow.
    // The "Save Changes" button is on the PARENT sheet (ServiceEditor).
    // The ToolEditor writes to the parent state.
    // But the tool execution relies on the backend having the tool registered.

    // This reveals a workflow issue: "Live Test" only works if the tool is ALREADY saved to backend.
    // But we are in the process of defining it.

    // So the workflow is:
    // 1. Define Tool.
    // 2. Close Tool Editor (auto-saves to parent state).
    // 3. Click "Save Changes" on Service Editor (persists to backend).
    // 4. Re-open Tool Editor.
    // 5. Click Execute.

    // Let's follow this flow.

    // Close Tool Editor Sheet
    await page.keyboard.press('Escape');
    await expect(page.getByRole('heading', { name: 'Edit get_uuid' })).not.toBeVisible();

    // Save Service
    await page.getByRole('button', { name: 'Save Changes' }).click();
    await expect(page.getByText('Service Updated')).toBeVisible();

    // Re-open Tool Editor
    // The tool should be listed in the manager now.
    // We need to find the "Edit" button for the tool row.
    // The row contains "get_uuid".
    const toolRow = page.locator('.space-y-4 .grid .flex', { hasText: 'get_uuid' }).first(); // Improved locator strategy
    // Or just look for the edit button inside the card that has 'get_uuid'
    // Actually, in `http-tool-manager.tsx`, it's a Card with `flex items-center justify-between`.
    // We can find the button by clicking the edit icon near 'get_uuid'.
    await page.getByRole('button', { name: 'Edit' }).click(); // There might be multiple if previous test left junk? No, fresh service.

    // Wait for Tool Editor to open
    await expect(page.getByText('Edit get_uuid')).toBeVisible();

    // Fill arguments again (state might be lost on close, which is expected)
    await page.getByLabel('Test Arguments (JSON)').fill('{"foo": "bar"}');

    // Click Execute
    await page.getByRole('button', { name: 'Execute' }).click();

    // Wait for result
    await expect(page.getByText('Execution Successful')).toBeVisible();

    // Verify Result Content
    // httpbin /uuid returns { uuid: "..." }
    await expect(page.locator('pre').filter({ hasText: 'uuid' })).toBeVisible();
  });
});
