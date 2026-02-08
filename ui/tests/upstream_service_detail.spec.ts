/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Upstream Service Detail Page', () => {
  const serviceName = 'e2e-detail-test-service';

  test.beforeAll(async ({ request }) => {
    // Seed the database with a test service
    const response = await request.post('/api/v1/services', {
      data: {
        id: serviceName,
        name: serviceName,
        version: "1.0.0",
        http_service: {
            address: "http://example.com"
        },
        priority: 10,
        disable: false
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    // Clean up
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should display ServiceEditor and save changes', async ({ page, request }) => {
    // 1. Navigate to the detail page
    await page.goto(`/upstream-services/${serviceName}`);

    // 2. Verify Page Title
    await expect(page.getByRole('heading', { level: 1 })).toContainText(serviceName);

    // 3. Go to Connection Tab
    await page.getByRole('tab', { name: 'Connection' }).click();

    // 4. Force selection of Service Type to ensure state is initialized correctly
    // This handles potential hydration issues where httpService might be missing initially
    const typeSelect = page.getByRole('combobox', { name: 'Service Type' });
    await expect(typeSelect).toBeVisible();

    // Check if HTTP Service is already selected or select it
    // Note: If value is already 'http', clicking the item might be redundant but safe.
    // However, we want to trigger the onChange logic to ensure httpService is populated.
    // If it's already "HTTP Service", handleTypeChange might not fire if Select doesn't emit on same value.
    // But getType() defaults to http.

    // We try to switch to something else and back, OR just fill the input if visible.
    // But since Base URL wasn't visible, we must trigger initialization.
    // Let's toggle to gRPC and back to HTTP.

    await typeSelect.click();
    await page.getByRole('option', { name: 'gRPC Service' }).click();

    await typeSelect.click();
    await page.getByRole('option', { name: 'HTTP Service' }).click();

    // 5. Modify the address
    const addressInput = page.getByLabel('Base URL');
    await expect(addressInput).toBeVisible();
    await addressInput.fill('http://example.com/updated');
    await addressInput.blur();

    // 6. Save Changes
    const responsePromise = page.waitForResponse(response =>
        response.url().includes(`/api/v1/services/${serviceName}`) &&
        response.request().method() === 'PUT'
    );

    const saveButton = page.getByRole('button', { name: 'Save Changes' });
    await expect(saveButton).toBeEnabled();
    await saveButton.click();

    // Wait for the response
    const response = await responsePromise;

    if (response.status() !== 200) {
        console.log("Save failed with status:", response.status());
        try {
             const body = await response.text();
             console.log("Response body:", body);
        } catch (e) {
            console.log("Could not read response body");
        }
    }

    expect(response.status()).toBe(200);

    // 7. Verify Persistence via API
    await expect(async () => {
        const response = await request.get(`/api/v1/services/${serviceName}`);
        expect(response.ok()).toBeTruthy();
        const service = await response.json();
        // Check for updated address
        const address = service.http_service?.address || service.httpService?.address;
        expect(address).toBe('http://example.com/updated');
    }).toPass({ timeout: 10000 });
  });
});
