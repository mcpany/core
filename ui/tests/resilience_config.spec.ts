/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resilience Configuration E2E', () => {
  const serviceName = 'e2e-resilience-test-service';

  test.beforeAll(async ({ request }) => {
    // Seed the database with a test service
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        httpService: {
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

  test('should configure resilience settings and persist them', async ({ page, request }) => {
    // 1. Navigate to the detail page
    await page.goto(`/upstream-services/${serviceName}`);

    // 2. Navigate to Settings tab -> Advanced tab
    // Note: The structure in UI is UpstreamServiceDetail -> ServiceEditor (which has tabs)
    // The "Settings" tab in `upstream_service_detail.spec.ts` refers to the tab in the DETAIL page that holds the editor.
    // Let's assume the editor is visible or we click "Settings" first.
    // In `upstream_service_detail.spec.ts`: await page.getByRole('tab', { name: 'Settings' }).click();
    await page.getByRole('tab', { name: 'Settings' }).click();

    // 3. Go to Advanced Tab
    await page.getByRole('tab', { name: 'Advanced' }).click();

    // 4. Verify Resilience Editor is present
    await expect(page.getByText('Global Timeout')).toBeVisible();
    await expect(page.getByText('Circuit Breaker')).toBeVisible();

    // 5. Configure Timeout
    const timeoutInput = page.getByLabel('Global Timeout');
    await timeoutInput.fill('45s');

    // 6. Configure Circuit Breaker
    const consecutiveFailuresInput = page.getByLabel('Consecutive Failures');
    await consecutiveFailuresInput.fill('10');

    const openDurationInput = page.getByLabel('Open Duration');
    await openDurationInput.fill('120s');

    // 7. Configure Retry Policy
    const retriesInput = page.getByLabel('Number of Retries');
    await retriesInput.fill('5');

    // 8. Save Changes
    const saveButton = page.getByRole('button', { name: 'Save Changes' });
    await saveButton.click();

    // 9. Verify Toast
    await expect(page.getByText('Service Updated').first()).toBeVisible();

    // 10. Verify Persistence via API
    const response = await request.get(`/api/v1/services/${serviceName}`);
    expect(response.ok()).toBeTruthy();
    const service = await response.json();

    // Verify Timeout
    expect(service.resilience).toBeDefined();
    // API returns snake_case for fields inside resilience usually, or whatever client.ts mapping logic produces?
    // Wait, `request.get` returns the raw JSON from backend.
    // Backend protobuf uses snake_case JSON mapping.
    // `ResilienceConfig` -> `timeout` (json_name="timeout")
    // `CircuitBreakerConfig` -> `consecutiveFailures` (json_name="consecutiveFailures")

    // The backend might return duration as string "45s".
    expect(service.resilience.timeout).toBe('45s');

    // Circuit Breaker
    expect(service.resilience.circuit_breaker).toBeDefined();
    expect(service.resilience.circuit_breaker.consecutiveFailures).toBe(10);
    expect(service.resilience.circuit_breaker.openDuration).toBe('120s');

    // Retry Policy
    expect(service.resilience.retry_policy).toBeDefined();
    expect(service.resilience.retry_policy.numberOfRetries).toBe(5);
  });
});
