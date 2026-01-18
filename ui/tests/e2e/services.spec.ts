/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Services Feature', () => {
  test.beforeEach(async ({ page, request }) => {
    // Seed data
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
        if (!r1.ok() && r1.status() !== 409) {
            console.error(`Failed to seed svc_01: ${r1.status()} ${await r1.text()}`);
        }
        expect(r1.status() === 200 || r1.status() === 201 || r1.status() === 409).toBeTruthy();

        const r2 = await request.post('/api/v1/services', {
            data: {
               id: "svc_02",
               name: "User Service",
               disable: false,
               version: "v1.0",
               grpc_service: { address: "localhost:50051", use_reflection: true, tools: [], resources: [] }
            }
        });
        if (!r2.ok() && r2.status() !== 409) {
            console.error(`Failed to seed svc_02: ${r2.status()} ${await r2.text()}`);
        }
        expect(r2.status() === 200 || r2.status() === 201 || r2.status() === 409).toBeTruthy();
    } catch (e) {
        console.log("Seeding interaction failed", e);
        throw e;
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

    // Select Custom Service template
    await page.getByText('Custom Service').click();

    const serviceName = `new-service-${Date.now()}`;
    await page.fill('input[id="name"]', serviceName);

    // Switch to Connection tab
    await page.getByRole('tab', { name: 'Connection' }).click();

    await page.getByRole('combobox').click();
    await page.getByRole('option', { name: 'HTTP' }).click();

    // With new editor, label is "Base URL" for HTTP
    const addressInput = page.getByPlaceholder('https://api.example.com');
    await expect(addressInput).toBeVisible();
    await addressInput.fill('http://localhost:8080');

    await page.getByRole('button', { name: 'Save Changes' }).click();

    // Explicitly wait for the service to appear in the list with a generous timeout
    await expect(page.getByText(serviceName)).toBeVisible({ timeout: 10000 });

    const newServiceRow = page.locator('tr').filter({ hasText: serviceName });
    await newServiceRow.waitFor({ state: 'visible', timeout: 5000 });
    await newServiceRow.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    // Check general tab which is default
    await expect(page.locator('input[id="name"]')).toHaveValue(serviceName);
    await page.getByRole('button', { name: 'Cancel' }).click();
  });
});
