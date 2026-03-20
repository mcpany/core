/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Bulk Service Actions', () => {

  const servicePrefix = `e2e-bulk-${Date.now()}`;
  const services = [
    { name: `${servicePrefix}-1`, tags: ["prod"] },
    { name: `${servicePrefix}-2`, tags: ["dev"] },
    { name: `${servicePrefix}-3`, tags: ["prod"] },
  ];

  test.beforeAll(async ({ request }) => {
    // Seed real services
    for (const svc of services) {
      const response = await request.post('/api/v1/services', {
        data: {
          name: svc.name,
          http_service: { address: "http://localhost:8001" },
          tags: svc.tags,
          priority: 10
        }
      });
      expect(response.ok()).toBeTruthy();
    }
  });

  test.afterAll(async ({ request }) => {
    // Clean up
    for (const svc of services) {
      await request.delete(`/api/v1/services/${svc.name}`).catch(() => {});
    }
  });

  test('should select all services and show bulk actions', async ({ page }) => {
    await page.goto('/upstream-services');

    // Wait for services to load by filtering
    await page.getByPlaceholder('Filter by tag...').fill('prod, dev');
    await expect(page.getByText(`${servicePrefix}-1`)).toBeVisible();
    await expect(page.getByText(`${servicePrefix}-2`)).toBeVisible();
    await expect(page.getByText(`${servicePrefix}-3`)).toBeVisible();

    // Check "Select All" checkbox using role
    const selectAllCheckbox = page.getByRole('checkbox', { name: 'Select all' });
    await selectAllCheckbox.check();

    // Verify bulk action buttons appear. Depending on what else is on the page, we might select more than 3.
    // But since we are filtering, it should be at least 3.
    await expect(page.getByText(/selected/)).toBeVisible();
    await expect(page.getByRole('button', { name: 'Enable' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Disable' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Delete' })).toBeVisible();
  });

  test('should select individual services', async ({ page }) => {
     await page.goto('/upstream-services');
     await expect(page.getByText(`${servicePrefix}-1`)).toBeVisible();

     // Select first service
     await page.getByRole('checkbox', { name: `Select ${servicePrefix}-1` }).check();

     // Verify 1 selected
     await expect(page.getByText('1 selected')).toBeVisible();

     // Select second service
     await page.getByRole('checkbox', { name: `Select ${servicePrefix}-2` }).check();
     await expect(page.getByText('2 selected')).toBeVisible();
  });

  test('should toggle services', async ({ page }) => {
      await page.goto('/upstream-services');
      await expect(page.getByText(`${servicePrefix}-1`)).toBeVisible();

      // Select service-1 and service-3
      await page.getByRole('checkbox', { name: `Select ${servicePrefix}-1` }).check();
      await page.getByRole('checkbox', { name: `Select ${servicePrefix}-3` }).check();

      // Click Disable
      await page.getByRole('button', { name: 'Disable' }).click();

      // Verify they are disabled. The UI should show "Disabled" badge or similar.
      // We can check if the status changed.
      await expect(page.getByText(`${servicePrefix}-1`).locator('..').getByText('Disabled')).toBeVisible({ timeout: 5000 });
      await expect(page.getByText(`${servicePrefix}-3`).locator('..').getByText('Disabled')).toBeVisible({ timeout: 5000 });

      // Clean up by re-enabling so we don't break other tests
      await page.getByRole('checkbox', { name: `Select ${servicePrefix}-1` }).check();
      await page.getByRole('checkbox', { name: `Select ${servicePrefix}-3` }).check();
      await page.getByRole('button', { name: 'Enable' }).click();
      await expect(page.getByText(`${servicePrefix}-1`).locator('..').getByText('Disabled')).toBeHidden({ timeout: 5000 });
  });

    test('should delete services', async ({ page }) => {
      // Handle confirm dialog
      page.on('dialog', dialog => dialog.accept());

      await page.goto('/upstream-services');
      await expect(page.getByText(`${servicePrefix}-1`)).toBeVisible();

      // Select service-2
      await page.getByRole('checkbox', { name: `Select ${servicePrefix}-2` }).check();

      // Click Delete
      await page.getByRole('button', { name: 'Delete' }).click();

      // Verify it's gone
      await expect(page.getByText(`${servicePrefix}-2`)).toBeHidden({ timeout: 5000 });
  });

});
