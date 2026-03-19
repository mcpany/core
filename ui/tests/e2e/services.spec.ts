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

  const extractLastPathSegment = (url: string) => decodeURIComponent(url.split('/').pop() || '');
  const extractServiceNameFromStatusUrl = (url: string) => {
    const match = url.match(/\/api\/v1\/services\/([^/]+)\/status$/);
    return match ? decodeURIComponent(match[1]) : '';
  };

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

    await page.route(url => /\/api\/v1\/services\/[^/]+$/.test(url.pathname), async route => {
        const serviceName = extractLastPathSegment(route.request().url());
        const service = services.find((candidate) => candidate.name === serviceName);
        if (!service) {
            await route.fulfill({ status: 404, json: { error: 'service not found' } });
            return;
        }

        await route.fulfill({ json: { service } });
    });

    await page.route(url => url.pathname.endsWith('/status'), async route => {
        const serviceName = extractServiceNameFromStatusUrl(route.request().url());
        const service = services.find((candidate) => candidate.name === serviceName);

        await route.fulfill({
            json: {
                tools: service?.tools ?? [],
            },
        });
    });

    await page.route(url => url.pathname.endsWith('/api/v1/dashboard/traffic'), async route => {
        await route.fulfill({ json: [] });
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
    await page.getByRole('link', { name: 'Payment Gateway' }).click();
    await expect(page.getByRole('heading', { name: 'Payment Gateway' })).toBeVisible();

    await page.getByRole('tab', { name: /Tools/ }).click();

    const toolCard = page.locator('[class*="grid"] > *').filter({ hasText: 'process_payment' }).first();
    await expect(toolCard).toContainText('Process a payment via Stripe.');
    await toolCard.getByRole('button', { name: 'View Schema' }).click();

    const dialog = page.getByRole('dialog');

    // SchemaViewer doesn't use table headers. We look for properties and descriptions directly.
    await expect(dialog.getByText('amount', { exact: true })).toBeVisible();
    await expect(dialog.getByText('currency', { exact: true })).toBeVisible();

    // SchemaViewer renders type badges with uppercase CSS, which can sometimes interfere with getByText
    // We'll check for the existence of the info icons which indicate descriptions are loaded
    // or just rely on the property names existing which confirms the tree rendered.
    const typeBadges = dialog.locator('span.font-mono.uppercase');
    await expect(typeBadges.first()).toBeVisible();
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
