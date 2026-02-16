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
        httpService: { address: "https://httpbin.org" }, // Use a real external service (reliable)
        priority: 0,
        loadBalancingStrategy: 0,
    };

    // For gRPC, we need a valid address.
    // If we don't have a real one, we can use a dummy one but expect it to be "down" or registered.
    // However, the test only checks list visibility.
    const userService: UpstreamServiceConfig = {
        name: "User Service",
        version: "v1.0",
        disable: false,
        // Using a dummy local address. Registration should succeed even if connection fails initially (depending on server strictness)
        // If server validates connectivity on registration, this might fail.
        // Assuming loose registration or it will be marked "down".
        grpcService: { address: "localhost:50052" },
        priority: 0,
        loadBalancingStrategy: 0,
    };

    await Seeder.registerService(paymentGateway);
    await Seeder.registerService(userService);

    await page.goto('/upstream-services');
  });

  test.afterEach(async () => {
      // Cleanup seeded data only!
      // To be safe in parallel environment, strictly we should only delete what we created.
      // But Seeder.cleanup deletes all.
      // For now, let's just delete the specific ones we created to avoid nuking others.
      try {
        await Seeder.registerService({ name: "Payment Gateway", id: "Payment Gateway" } as any).then(() => {}, () => {}); // creating dummy to delete? No.
        // Seeder doesn't have delete specific.
        // Let's rely on Seeder.cleanup() for this specific test suite if it runs in isolation.
        // But for better robustness:
        await apiClient.unregisterService("Payment Gateway").catch(() => {});
        await apiClient.unregisterService("User Service").catch(() => {});
      } catch (e) {
          // ignore
      }
  });

  // Need to import apiClient for manual cleanup
  const { apiClient } = require('@/lib/client');


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

    // Clean up the new service
    await apiClient.unregisterService(serviceName).catch(() => {});
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
