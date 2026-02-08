/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('HTTP Tool Editor', () => {
  const serviceName = 'e2e-http-tool-test';

  test.beforeAll(async ({ request }) => {
    // Seed HTTP service
    // Ensure cleanup first just in case
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});

    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        http_service: {
            address: "http://example.com"
        }
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should define tools for HTTP service', async ({ page, request }) => {
    // Navigate to upstream services
    await page.goto('/upstream-services');

    // Wait for list to load
    await expect(page.getByText(serviceName)).toBeVisible();

    // Find row, click edit.
    const row = page.getByRole('row', { name: serviceName });
    await row.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    // Wait for Sheet to open.
    await expect(page.getByRole('dialog', { name: 'Edit Service' })).toBeVisible();

    // Click "Tools" tab.
    await page.getByRole('tab', { name: 'Tools' }).click();

    // Click "Add Tool"
    await page.getByRole('button', { name: 'Add Tool' }).click();

    // Sheet for Tool Editor should open (it's a nested sheet or just covers content?)
    // In HttpToolManager, it uses Sheet. Sheet inside Sheet works in shadcn/radix if handled well, or it replaces?
    // Let's assume it opens.
    await expect(page.getByText('Edit new_tool')).toBeVisible();

    // Fill details
    await page.getByLabel('Tool Name').fill('get_weather');
    await page.getByLabel('Description').fill('Get weather info');

    // Call details
    await page.getByLabel('Endpoint Path').fill('/weather');

    // Add Parameter
    await page.getByRole('button', { name: 'Add Parameter' }).click();
    await page.getByLabel('Name', { exact: true }).fill('city');

    // Close Tool Editor Sheet
    // There isn't a "Done" button in my implementation, just auto-save to parent state.
    // So we close the sheet. Pressing Escape might close the top-most sheet.
    await page.keyboard.press('Escape');
    await expect(page.getByRole('heading', { name: 'Edit get_weather' })).not.toBeVisible();

    // Verify tool is listed in Manager
    // Use first() if strict mode is still an issue (though closing sheet should fix it),
    // or be more specific. The sheet title should be gone/hidden now.
    await expect(page.getByText('get_weather', { exact: true })).toBeVisible();
    await expect(page.getByText('/weather')).toBeVisible();

    // Save Service
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // Verify Toast (Specific locator to avoid multiple matches)
    await expect(page.getByRole('status').filter({ hasText: 'Service Updated' })).toBeVisible();

    // Verify Persistence via API
    const response = await request.get(`/api/v1/services/${serviceName}`);
    expect(response.ok()).toBeTruthy();
    const service = await response.json();
    const httpService = service.http_service || service.httpService;

    expect(httpService.tools).toHaveLength(1);
    expect(httpService.tools[0].name).toBe('get_weather');

    // Verify Call Definition
    const callId = httpService.tools[0].callId || httpService.tools[0].call_id;
    const calls = httpService.calls; // Map

    // Check if calls is populated
    expect(calls).toBeDefined();
    const call = calls[callId];
    expect(call).toBeDefined();

    const endpointPath = call.endpointPath || call.endpoint_path;
    expect(endpointPath).toBe('/weather');
  });
});
