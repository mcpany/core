/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Services Feature', () => {

  test.beforeEach(async ({ page, request }) => {
    // Seed global state directly, rather than mocking page.route
    await seedGlobalState(request);

    await page.goto('/upstream-services');
  });

  test('should list services, allow toggle, and manage services', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Services');

    // Verify services are listed
    await expect(page.getByText('Payment Gateway').first()).toBeVisible();
    await expect(page.getByText('User Service').first()).toBeVisible();

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
    // Click on the link to open details instead of row
    await page.getByRole('link', { name: 'Payment Gateway', exact: true }).click();

    // Tools are now in the General tab by default
    await expect(page.getByRole('tab', { name: 'Tools 1' }).or(page.getByRole('tab', { name: 'Tools' })).first()).toBeVisible();
    await page.getByRole('tab', { name: /Tools.*/ }).click();

    // Click View Schema button
    await page.locator('button[title="View Schema"]').click();

    // The dialog should appear and it should have the visualizer table
    // we added SchemaVisualizer which renders a Table with headers "Property", "Type", "Description"
    await expect(page.getByRole('dialog').getByRole('columnheader', { name: 'Property' })).toBeVisible();
    await expect(page.getByRole('dialog').getByRole('columnheader', { name: 'Type' })).toBeVisible();

    // Should see the properties we defined
    await expect(page.getByRole('dialog').getByText('amount', { exact: true })).toBeVisible();
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
