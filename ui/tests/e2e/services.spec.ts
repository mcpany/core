/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Services Feature', () => {
  test.beforeEach(async ({ page, request }) => {
    await seedGlobalState(request);
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
