/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { Seeder } from '../utils/seeder';

test.describe('Services Feature', () => {
  const seeder = new Seeder();

  test.beforeEach(async ({ page }) => {
    // Seed database with real data
    await seeder.registerService({
        id: "payment-gateway",
        name: "Payment Gateway",
        version: "v1.2.0",
        disable: false,
        httpService: { address: "https://stripe.com" },
        priority: 0,
        loadBalancingStrategy: 0,
    });
    await seeder.registerService({
        id: "user-service",
        name: "User Service",
        version: "v1.0",
        disable: false,
        grpcService: { address: "localhost:50051" },
        priority: 0,
        loadBalancingStrategy: 0,
    });

    await page.goto('/upstream-services');
  });

  test.afterEach(async () => {
    await seeder.cleanup();
  });

  test('should list services, allow toggle, and manage services', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Upstream Services');

    // Verify services are listed
    await expect(page.getByText('Payment Gateway')).toBeVisible();
    await expect(page.getByText('User Service')).toBeVisible();

    // Verify Toggle exists and is interactive
    // Note: Use a more specific selector if possible to avoid ambiguity
    const paymentRow = page.locator('tr').filter({ hasText: 'Payment Gateway' });
    const switchBtn = paymentRow.getByRole('switch');
    await expect(switchBtn).toBeVisible();
    // Toggle off
    await switchBtn.click();
    // We can verify the state changed if the UI updates optimistically or fetches.
    // The real backend will update the status.

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

    // Ensure we clean up this manually created service too
    seeder.trackService(serviceName);
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
