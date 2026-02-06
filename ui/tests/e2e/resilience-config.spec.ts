/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Service Resilience Configuration (Real Data)', () => {
  const serviceName = `resilient-service-${Date.now()}`;

  test.beforeEach(async ({ request }) => {
    // Seed the database with a service using real API
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        http_service: { address: "https://api.example.com" },
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
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterEach(async ({ request }) => {
      // Clean up
      await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should display and edit resilience configuration', async ({ page, request }) => {
    await page.goto('/upstream-services');

    // 1. Open the service editor
    const row = page.locator('tr').filter({ hasText: serviceName });
    await expect(row).toBeVisible();
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

    // 6. Verify Backend State via API
    const response = await request.get(`/api/v1/services/${serviceName}`);
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    const service = data.service || data; // Handle both wrapper formats if any

    expect(service.resilience.timeout).toBe('60s');
    expect(service.resilience.retry_policy.number_of_retries).toBe(5);
    // Note: Proto might return floats slightly differently or as numbers, Playwright expect handles it usually
    expect(service.resilience.circuit_breaker.failure_rate_threshold).toBe(0.2);
  });
});
