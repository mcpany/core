/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Onboarding Experience', () => {
  test.beforeEach(async ({ request }) => {
    // Clean up all services to ensure empty state
    // We try to list services. If it fails (e.g. backend down), test will fail, which is good.
    const response = await request.get('/api/v1/services');
    if (!response.ok()) {
        console.error("Failed to list services during cleanup");
        return;
    }
    const data = await response.json();
    const services = Array.isArray(data) ? data : (data.services || []);

    for (const service of services) {
        // Use name if id is missing
        const id = service.name;
        if (id) {
            await request.delete(`/api/v1/services/${id}`);
        }
    }
  });

  test('should show onboarding hero when no services exist', async ({ page }) => {
    await page.goto('/');

    // Verify Hero is visible
    await expect(page.getByText('Welcome to MCP Any')).toBeVisible();
    await expect(page.getByText('Connect Service')).toBeVisible();

    // Verify Dashboard Header is NOT visible
    await expect(page.getByRole('heading', { level: 1, name: 'Dashboard' })).not.toBeVisible();
  });

  test('should show dashboard when services exist', async ({ page, request }) => {
    // Seed a service
    const seed = await request.post('/api/v1/services', {
      data: {
        name: 'test-service-onboarding',
        http_service: { address: 'http://example.com' },
        disable: false
      }
    });
    expect(seed.ok()).toBeTruthy();

    await page.goto('/');

    // Verify Dashboard Header IS visible
    await expect(page.getByRole('heading', { level: 1, name: 'Dashboard' })).toBeVisible();

    // Verify Hero is NOT visible
    await expect(page.getByText('Welcome to MCP Any')).not.toBeVisible();
  });
});
