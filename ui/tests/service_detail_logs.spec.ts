/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from './e2e/test-data';

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

  test('should display Logs tab and LogStream component', async ({ page, request }) => {
    // Login first
    await seedUser(request, "e2e-admin");
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', 'e2e-admin');
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/', { timeout: 15000 });

    // 1. Navigate to the detail page
    // Use the actual ID returned from the backend creation
    await page.goto(`/service/${serviceId}`);

    // 2. Verify Page Title to ensure we loaded
    // Depending on loading state, title might take a moment
    await expect(page.getByRole('heading', { level: 3 })).toContainText(serviceName);

    // 3. Click Logs Tab
    const logsTab = page.getByRole('tab', { name: 'Logs' });
    await expect(logsTab).toBeVisible();
    await logsTab.click();

    // 4. Verify LogStream is visible
    // "Live Logs" is the h1 in LogStream
    await expect(page.getByText('Live Logs')).toBeVisible();

    await cleanupUser(request, "e2e-admin");
  });
});
