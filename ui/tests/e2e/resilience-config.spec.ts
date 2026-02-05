/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Service Resilience Configuration', () => {
  // "Database"
  interface MockService {
      name: string;
      id: string;
      [key: string]: any;
  }
  const services: MockService[] = [
    {
        name: "Resilient Service",
        id: "resilient-service-123",
        type: "http",
        http_service: { address: "https://api.example.com" },
        status: "up",
        version: "v1.0.0",
        enabled: true,
        resilience: {
            timeout: "30s",
            retry_policy: {
                number_of_retries: 3,
                base_backoff: "100ms",
                max_backoff: "1s"
            },
            circuit_breaker: {
                failure_rate_threshold: 0.5,
                consecutive_failures: 5,
                open_duration: "60s",
                half_open_requests: 3
            }
        }
    }
  ];

  test.beforeEach(async ({ page }) => {
    // Mock registration API
    await page.route(url => url.pathname.includes('/api/v1/services'), async route => {
        const method = route.request().method();
        const url = route.request().url();

        if (method === 'GET') {
            if (url.endsWith('/api/v1/services')) {
                // List
                await route.fulfill({ json: { services } });
            } else {
                // Get one (by name or ID) - simplified matching
                const idOrName = url.split('/').pop();
                const service = services.find(s => s.name === idOrName || s.id === idOrName);
                if (service) {
                    await route.fulfill({ json: { service } });
                } else {
                    await route.fulfill({ status: 404 });
                }
            }
        } else if (method === 'PUT') {
            // Update
            const body = route.request().postDataJSON();
            const idx = services.findIndex(s => s.name === body.name || s.id === body.id);
            if (idx !== -1) {
                services[idx] = { ...services[idx], ...body };
                await route.fulfill({ json: services[idx] });
            } else {
                await route.fulfill({ status: 404 });
            }
        } else if (method === 'POST' && url.endsWith('/validate')) {
             await route.fulfill({ json: { valid: true } });
        } else {
            await route.continue();
        }
    });

    await page.goto('/upstream-services');
  });

  test('should display and edit resilience configuration', async ({ page }) => {
    // 1. Open the service editor
    const row = page.locator('tr').filter({ hasText: 'Resilient Service' });
    await row.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    await expect(page.getByRole('dialog')).toBeVisible();

    // 2. Navigate to Advanced tab
    await page.getByRole('tab', { name: 'Advanced' }).click();

    // 3. Verify seeded values are displayed
    await expect(page.getByLabel('Timeout')).toHaveValue('30s');
    await expect(page.getByLabel('Max Retries')).toHaveValue('3');
    await expect(page.getByLabel('Base Backoff')).toHaveValue('100ms');
    await expect(page.getByLabel('Failure Rate Threshold')).toHaveValue('0.5');

    // 4. Edit values
    await page.getByLabel('Timeout').fill('60s');
    await page.getByLabel('Max Retries').fill('5');
    await page.getByLabel('Failure Rate Threshold').fill('0.2');

    // 5. Save
    await page.getByRole('button', { name: 'Save Changes' }).click();
    await expect(page.getByRole('dialog')).toBeHidden();

    // 6. Verify "Database" state (Backend verification)
    // In a real E2E, we would query the API. Here we check our mock DB variable.
    // However, Playwright runs in a separate process, so we can't access `services` variable directly from here
    // if the route handler was running in the browser context?
    // Wait, `page.route` handler runs in the Node.js context of the test runner!
    // So `services` variable IS accessible and shared!

    const updatedService = services.find(s => s.name === "Resilient Service");
    expect(updatedService.resilience.timeout).toBe('60s');
    expect(updatedService.resilience.retry_policy.number_of_retries).toBe(5);
    expect(updatedService.resilience.circuit_breaker.failure_rate_threshold).toBe(0.2);
  });
});
