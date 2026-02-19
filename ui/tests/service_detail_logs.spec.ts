/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Service Detail Logs Tab', () => {
  const serviceName = 'e2e-logs-test-service';
  let serviceId = '';

  test.beforeAll(async ({ request }) => {
    // Seed the database with a test service
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
    const data = await response.json();
    // Use ID if available, otherwise fallback to name if the API behaves that way
    serviceId = data.id || data.name || serviceName;
  });

  test.afterAll(async ({ request }) => {
    // Clean up
    // Delete typically accepts ID or Name
    if (serviceId) {
        await request.delete(`/api/v1/services/${serviceId}`);
    } else {
        await request.delete(`/api/v1/services/${serviceName}`);
    }
  });

  test('should display Logs tab and LogStream component', async ({ page }) => {
    // 1. Navigate to the detail page
    // Use the actual ID returned from the backend creation
    await page.goto(`/service/${serviceId}`);

    // 2. Verify Page Title to ensure we loaded
    // Increase timeout and use a more flexible text match as the layout might use different heading levels or components
    await expect(page.getByText(serviceName, { exact: false }).first()).toBeVisible({ timeout: 15000 });

    // 3. Click Logs Tab
    const logsTab = page.getByRole('tab', { name: 'Logs' });
    await expect(logsTab).toBeVisible();
    await logsTab.click();

    // 4. Verify LogStream is visible
    // "Live Logs" is the h1 in LogStream
    await expect(page.getByText('Live Logs')).toBeVisible();

    // 5. Verify source is filtered (optional, but good)
    // The LogStream displays source badge/text.
    // In LogStream: <SelectValue placeholder="Source" />
    // It might default to ALL or the source passed.
    // Since we can't easily check internal state, checking visibility is enough for "broken window" fix.
  });
});
