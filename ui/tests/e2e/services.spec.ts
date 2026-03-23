/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

<<<<<<< HEAD
test.describe.skip('Services Feature', () => {
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




=======
test.describe('Services Feature', () => {
  test.beforeEach(async ({ page, request }) => {
    await seedGlobalState(request);
>>>>>>> 7ae9f1781 (test(ui): Refactor E2E locators and assertions for schema visualizer)
    await page.goto('/upstream-services');
  });

  test.skip('should list services, allow toggle, and manage services', async ({ page }) => {
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

    // Select Custom Service template (actually empty template logic if applicable)
    // Wait for the Template selection view to be visible
    await expect(page.getByText('Select Service Template')).toBeVisible();

    // In RegisterServiceDialog, there is a template selector.
    // Assuming there's a way to start from scratch or pick HTTP custom.
    // If there is a "Custom Service" or "Blank HTTP Service" option, click it.
    // Otherwise, we might need to adjust based on the actual templates rendered.
    // Let's assume 'Blank HTTP Service' or similar exists, or we can just click 'HTTP' if it's there.
    // Based on standard implementation, there's usually a "Custom HTTP" option.
    const customHttpOption = page.locator('text=Custom HTTP').first();
    if (await customHttpOption.isVisible()) {
        await customHttpOption.click();
    } else {
        // Fallback: If no template selector blocks us, or if we can just proceed
        // Try clicking a generic "Custom" or "Blank"
        const customOption = page.locator('text=Custom').first();
        if (await customOption.isVisible()) {
             await customOption.click();
        }
    }

    // Now we should be in the form view
    await expect(page.getByText('Configure Service')).toBeVisible();

    const serviceName = `new-service-${Date.now()}`;
    await page.fill('input[placeholder="my-service"]', serviceName);

    // Protocol selection is now a select dropdown named 'type'
    await page.locator('button[role="combobox"]').first().click();
    await page.getByRole('option', { name: 'HTTP' }).click();

    const addressInput = page.getByPlaceholder('https://api.example.com');
    await expect(addressInput).toBeVisible();
    await addressInput.fill('http://localhost:8080');

    await page.getByRole('button', { name: 'Register Service' }).click();
    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10000 });

    // Should be visible in the list now
    await expect(page.getByRole('link', { name: serviceName })).toBeVisible({ timeout: 10000 });

    const newServiceRow = page.locator('tr').filter({ hasText: serviceName });
    await newServiceRow.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    // The editor sheet uses id="name"
    await expect(page.locator('input[id="name"]')).toHaveValue(serviceName);
    await page.getByRole('button', { name: 'Cancel' }).click();
  });

<<<<<<< HEAD
  test.skip('should render schema visualizer in service tools dialog', async ({ page }) => {
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
=======
  test('should render schema visualizer in service tools dialog', async ({ page }) => {
    // Navigate straight to the specific service page
    await page.goto('/upstream-services/Payment%20Gateway');
    await page.waitForLoadState('networkidle');

    // Wait for the service page to render
    await expect(page.locator('h1').filter({ hasText: 'Payment Gateway' })).toBeVisible({ timeout: 10000 });

    // Switch to the Tools tab
    const toolsTab = page.locator('[role="tab"], button').filter({ hasText: 'Tools' }).first();
    await toolsTab.click();

    // Look for process_payment tool card
    const processPaymentTitle = page.locator('div, span, p, h1, h2, h3, h4').filter({ hasText: 'process_payment' }).first();
    await expect(processPaymentTitle).toBeVisible({ timeout: 10000 });

    // Click the specific View Schema button for this tool.
    const viewSchemaBtn = page.locator('button[title="View Schema"], .lucide-file-json').first();
    await viewSchemaBtn.click();

    // Validate the Schema Visualizer opens and displays the seeded input schema properties
    await expect(page.getByText('amount').first()).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Payment amount in cents').first()).toBeVisible();
    await expect(page.getByText('currency').first()).toBeVisible();
    await expect(page.getByText('3-letter ISO currency code').first()).toBeVisible();
>>>>>>> 7ae9f1781 (test(ui): Refactor E2E locators and assertions for schema visualizer)
  });

  test.skip('should navigate to logs from service list', async ({ page }) => {
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
