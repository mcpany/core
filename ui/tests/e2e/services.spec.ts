/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Services Feature', () => {
  test.beforeEach(async ({ page, request }) => {
    // Seed data
    page.on('console', msg => console.log(`BROWSER LOG: ${msg.text()}`));
    page.on('requestfailed', request => console.log(`BROWSER REQ FAILED: ${request.url()} ${request.failure()?.errorText}`));
    try {
        const r1 = await request.post('/api/v1/services', {
            data: {
                id: "svc_01",
                name: "Payment Gateway",
                connection_pool: { max_connections: 100 },
                disable: false,
                version: "v1.2.0",
                http_service: { address: "https://stripe.com", tools: [], resources: [] }
            }
        });
        if (!r1.ok() && r1.status() !== 409) console.error("Failed to seed Payment Gateway:", r1.status(), await r1.text());
        // expect(r1.ok()).toBeTruthy(); // Don't fail if 409

        const r2 = await request.post('/api/v1/services', {
            data: {
               id: "svc_02",
               name: "User Service",
               disable: false,
               version: "v1.0",
               http_service: { address: "http://http-echo-server:8080", tools: [], resources: [] }
            }
        });
        if (!r2.ok() && r2.status() !== 409) console.error("Failed to seed User Service:", r2.status(), await r2.text());
        // expect(r2.ok()).toBeTruthy();
    } catch (e) {
        console.log("Seeding failed or services already exist", e);
        // Ignore failure if it's just "already works" (we rely on list check)
        // ideally we check if it exists, but ignoring 409/failure for now is robust for retries
    }

    await page.goto('/services');
  });

  test('should list services, allow toggle, and manage services', async ({ page }) => {
    await expect(page.locator('h2')).toContainText('Services');

    // Verify services are listed
    await expect(page.getByText('Payment Gateway')).toBeVisible();
    await expect(page.getByText('User Service')).toBeVisible();

    // Verify Toggle exists and is interactive
    const paymentRow = page.locator('tr').filter({ hasText: 'Payment Gateway' });
    const switchBtn = paymentRow.getByRole('switch');
    await expect(switchBtn).toBeVisible();
    await switchBtn.click();

    // Register a new service
    await page.getByRole('button', { name: 'Add Service' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    const serviceName = `new-service-${Date.now()}`;
    await page.fill('input[id="name"]', serviceName);

    await page.getByRole('combobox').click();
    await page.getByRole('option', { name: 'HTTP' }).click();

    const addressInput = page.getByLabel('Endpoint');
    await expect(addressInput).toBeVisible();
    await addressInput.fill('http://localhost:8080');

    await page.getByRole('button', { name: 'Save Changes' }).click();
    await expect(page.getByText(serviceName)).toBeVisible();

    const newServiceRow = page.locator('tr').filter({ hasText: serviceName });
    await newServiceRow.getByRole('button', { name: 'Edit' }).click();
    await expect(page.locator('input[id="name"]')).toHaveValue(serviceName);
    await page.getByRole('button', { name: 'Cancel' }).click();
  });
});
