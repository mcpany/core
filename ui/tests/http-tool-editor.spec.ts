/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('HTTP Tool Editor', () => {
  const serviceName = 'e2e-http-tool-test';

  test.beforeAll(async ({ request }) => {
    // Seed the database with a test service
    // Ensure cleanup first just in case
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});

    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        http_service: {
            address: "http://example.com"
        },
        priority: 10
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    // Clean up
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should add and save a tool definition', async ({ page, request }) => {
    // 1. Navigate to the detail page
    await page.goto(`/upstream-services/${serviceName}`);

    // 2. Click "Tools" tab
    await page.getByRole('tab', { name: 'Tools' }).click();

    // 3. Click "Add Tool"
    await page.getByRole('button', { name: 'Add Tool' }).click();

    // 4. Fill Tool Details
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    await dialog.getByLabel('Name').fill('get_users');
    await dialog.getByLabel('Description').fill('Fetch list of users');

    // 5. Fill HTTP Request
    await dialog.getByLabel('Endpoint Path').fill('/users');
    // Default method is GET, which is what we want

    // 6. Save Tool
    await dialog.getByRole('button', { name: 'Save Tool' }).click();

    // 7. Verify Tool is listed
    await expect(page.getByText('get_users')).toBeVisible();
    await expect(page.getByText('GET', { exact: true })).toBeVisible(); // Exact match to avoid "GET" in other places if any
    await expect(page.getByText('/users')).toBeVisible();

    // 8. Save Service
    await page.getByRole('button', { name: 'Save Changes' }).click();
    await expect(page.getByText('Service Updated').first()).toBeVisible();

    // 9. Verify Persistence via API
    const response = await request.get(`/api/v1/services/${serviceName}`);
    expect(response.ok()).toBeTruthy();
    const body = await response.json();

    // The structure might be body.service or just body depending on the endpoint response
    const service = body.service || body;
    console.log('Service Body:', JSON.stringify(service, null, 2));

    // Check if tools are present
    const httpService = service.http_service;
    expect(httpService).toBeDefined();

    // Tools might be null if empty, but we added one.
    expect(httpService.tools).toHaveLength(1);
    expect(httpService.tools[0].name).toBe('get_users');

    // Check calls
    expect(httpService.calls).toBeDefined();
    const callIds = Object.keys(httpService.calls);
    expect(callIds).toHaveLength(1);

    const call = httpService.calls[callIds[0]];
    expect(call.endpoint_path).toBe('/users');
  });
});
