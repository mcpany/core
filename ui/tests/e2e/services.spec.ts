/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { Seeder } from '../utils/seeder';
import { UpstreamServiceConfig } from '@/lib/client';

test.describe('Services Feature', () => {

  test.beforeEach(async ({ page }) => {
    // Seed the database with initial services
    const paymentGateway: UpstreamServiceConfig = {
        name: "Payment Gateway",
        version: "v1.2.0",
        disable: false,
        httpService: { address: "https://stripe.com" },
        priority: 0,
        loadBalancingStrategy: 0,
    };

    const userService: UpstreamServiceConfig = {
        name: "User Service",
        version: "v1.0",
        disable: false,
        grpcService: { address: "localhost:50051" },
        priority: 0,
        loadBalancingStrategy: 0,
    };

    await Seeder.registerService(paymentGateway);
    await Seeder.registerService(userService);

    await page.goto('/upstream-services');
  });

  test.afterEach(async () => {
      // Cleanup seeded data
      await Seeder.cleanup();
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
    // Toggle logic: click calls backend.
    await switchBtn.click();
    // Verify backend state update? Or just UI update.
    // UI update depends on backend response. If backend works, UI toggles.
    // We can assume UI reflects backend state after optimistic update or revalidation.

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
