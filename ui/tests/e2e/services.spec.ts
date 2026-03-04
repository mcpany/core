/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Services Feature', () => {
  const services: any[] = [
    {
        name: "Payment Gateway",
        type: "http",
        address: "https://stripe.com",
        status: "up",
        version: "v1.2.0",
        enabled: true,
        tools: [{
            name: "process_payment",
            description: "Process a payment via Stripe.",
            inputSchema: {
                type: "object",
                properties: {
                    amount: {
                        type: "number",
                        description: "Payment amount in cents"
                    },
                    currency: {
                        type: "string",
                        description: "Currency code (e.g., USD)"
                    }
                },
                required: ["amount", "currency"]
            }
        }]
    },
    {
        name: "User Service",
        type: "grpc",
        address: "localhost:50051",
        status: "up",
        version: "v1.0",
        enabled: true
    }
  ];

  test.beforeEach(async ({ page }) => {
    // page.on('request', request => console.log('>>', request.method(), request.url()));

    // Mock registration API with dynamic state
    await page.route(url => url.pathname.endsWith('/api/v1/services'), async route => {
        const method = route.request().method();
        if (method === 'GET') {
            await route.fulfill({ json: { services } });
        } else if (method === 'POST') {
            const newSvc = route.request().postDataJSON();
            const created = { ...newSvc, status: 'up', enabled: true };
            services.push(created);
            await route.fulfill({ json: created });
        } else {
            await route.continue();
        }
    });

    await page.goto('/upstream-services');
  });

  test('should list services, allow toggle, and manage services', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Services');

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

    const addressInput = page.getByPlaceholder('https://api.example.com');
    await expect(addressInput).toBeVisible();
    await addressInput.fill('http://localhost:8080');

    await page.getByRole('button', { name: 'Save Changes' }).click();
    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10000 });

    // Should be visible in the list now
    await expect(page.getByText(serviceName)).toBeVisible({ timeout: 10000 });

    const newServiceRow = page.locator('tr').filter({ hasText: serviceName });
    await newServiceRow.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    await expect(page.locator('input[id="name"]')).toHaveValue(serviceName);
    await page.getByRole('button', { name: 'Cancel' }).click();
  });

  test('should render schema visualizer in service tools dialog', async ({ page }) => {
    const paymentRow = page.locator('tr').filter({ hasText: 'Payment Gateway' });

    // Click on the row to open details
    await paymentRow.click();

    // Click on tools tab
    await page.getByRole('tab', { name: 'Tools' }).click();

    // Should see process_payment tool
    await expect(page.getByText('process_payment')).toBeVisible();

    // Click View Schema button
    await page.locator('button[title="View Schema"]').click();

    // The dialog should appear and it should have the visualizer table
    // we added SchemaVisualizer which renders a Table with headers "Property", "Type", "Description"
    await expect(page.getByRole('dialog').getByRole('columnheader', { name: 'Property' })).toBeVisible();
    await expect(page.getByRole('dialog').getByRole('columnheader', { name: 'Type' })).toBeVisible();

    // Should see the properties we defined
    await expect(page.getByRole('dialog').getByText('amount')).toBeVisible();
    await expect(page.getByRole('dialog').getByText('currency')).toBeVisible();
    await expect(page.getByRole('dialog').getByText('Payment amount in cents')).toBeVisible();
  });

  test('should navigate to logs from service list', async ({ page }) => {
    const serviceName = 'Payment Gateway';
    const row = page.locator('tr').filter({ hasText: serviceName });

    // Open menu
    await row.getByRole('button', { name: 'Open menu' }).click();

    // Check View Logs link
    const viewLogsLink = page.getByRole('menuitem', { name: 'View Logs' });
    await expect(viewLogsLink).toBeVisible();

    // Click and verify navigation
    await viewLogsLink.click();

    // Should navigate to logs page with query param
    await expect(page).toHaveURL(/.*\/logs.*source=Payment/);
  });
});
