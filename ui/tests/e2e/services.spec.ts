/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Services Feature', () => {
  test.beforeEach(async ({ request, page }) => {
    await seedGlobalState(request);
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', 'e2e-admin-core');
    await page.fill('input[name="password"]', 'password');
    await Promise.all([
      page.waitForURL('/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
    ]);
    await expect(page).toHaveURL('/', { timeout: 15000 });
    await page.goto('/upstream-services');
  });

  test('should display seeded tags and filter correctly', async ({ page }) => {
      // Check if "Payment Gateway" is visible with its tag
      const paymentRow = page.locator('tr', { hasText: 'Payment Gateway' });
      await expect(paymentRow).toBeVisible();
      await expect(paymentRow.locator('text=financial')).toBeVisible();

      const userRow = page.locator('tr', { hasText: 'User Service' });
      await expect(userRow).toBeVisible();
      await expect(userRow.locator('text=core')).toBeVisible();

      // Filter by 'financial'
      await page.getByPlaceholder('Filter by tag...').fill('financial');

      // Verify Payment Gateway is still visible
      await expect(page.locator('tr', { hasText: 'Payment Gateway' })).toBeVisible();
      // Verify User Service is hidden
      await expect(page.locator('tr', { hasText: 'User Service' })).not.toBeVisible();
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

    // Select Custom Service template which maps to id="empty" in templates.ts
    const customOption = page.locator('h3').filter({ hasText: /^Custom Service$/ }).first();
    if (await customOption.isVisible()) {
         await customOption.click();
    }

    // Give form time to render
    await page.waitForTimeout(500);

    const serviceName = `new-service-${Date.now()}`;
    // Using a more generic selector logic that fits the editor since exact form varies
    const nameInput = page.locator('input[name="name"], input[placeholder="my-service"]').first();
    await expect(nameInput).toBeVisible({ timeout: 10000 });
    await nameInput.fill(serviceName);

    const addressInput = page.locator('input[name="address"], input[placeholder*="example.com"]').first();
    if (await addressInput.isVisible()) {
        await addressInput.fill('http://localhost:8080');
    }

    const saveBtn = page.locator('button').filter({ hasText: /Register Service|Save/i }).first();
    await saveBtn.click();
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

  test('should render schema visualizer in service tools dialog', async ({ page }) => {
    await page.getByRole('link', { name: 'Payment Gateway' }).click();
    await expect(page.getByRole('heading', { name: 'Payment Gateway' })).toBeVisible();

    await page.getByRole('tab', { name: /Tools/ }).click();

    const toolCard = page.locator('[class*="grid"] > *').filter({ hasText: 'process_payment' }).first();
    await expect(toolCard).toContainText('Process a payment');
    await toolCard.getByRole('button', { name: 'View Schema' }).click();

    const dialog = page.getByRole('dialog');

    // SchemaViewer doesn't use table headers. We look for properties directly.
    // Use a more relaxed selector to find the text since it might be wrapped in spans with other elements
    await expect(dialog.locator('span', { hasText: 'amount' }).first()).toBeVisible();
    await expect(dialog.locator('span', { hasText: 'currency' }).first()).toBeVisible();

    // Check for the existence of the type badges
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
