/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, login } from './test-data';

test.describe('Bulk Service Actions', () => {

  test.beforeEach(async ({ page, request }) => {
    const services = [
        { id: "s1", name: "service-1", http_service: { address: "http://example.com/8001" }, disable: false, tags: ["prod"] },
        { id: "s2", name: "service-2", http_service: { address: "http://example.com/8002" }, disable: true, tags: ["dev"] },
        { id: "s3", name: "service-3", http_service: { address: "http://example.com/8003" }, disable: false, tags: ["prod"] }
    ];
    await seedServices(request, services);
    await login(page);
  });

  test('should select all services and show bulk actions', async ({ page }) => {
    await page.goto('/upstream-services');

    // Wait for services to load
    await expect(page.getByText('service-1')).toBeVisible();

    // Check "Select All" checkbox using role
    const selectAllCheckbox = page.getByRole('checkbox', { name: 'Select all' });
    await selectAllCheckbox.check();

    // Verify bulk action buttons appear
    await expect(page.getByText('3 selected')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Enable' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Disable' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Delete' })).toBeVisible();
  });

  test('should select individual services', async ({ page }) => {
     await page.goto('/upstream-services');
     await expect(page.getByText('service-1')).toBeVisible();

     // Select first service
     await page.getByRole('checkbox', { name: 'Select service-1' }).check();

     // Verify 1 selected
     await expect(page.getByText('1 selected')).toBeVisible();

     // Select second service
     await page.getByRole('checkbox', { name: 'Select service-2' }).check();
     await expect(page.getByText('2 selected')).toBeVisible();
  });

  test('should toggle services', async ({ page }) => {
      await page.goto('/upstream-services');
      await expect(page.getByText('service-1')).toBeVisible();

      // Select service-1 and service-3
      await page.getByRole('checkbox', { name: 'Select service-1' }).check();
      await page.getByRole('checkbox', { name: 'Select service-3' }).check();

      // Setup response wait
      const p1 = page.waitForResponse(resp => resp.url().includes('service-1') && resp.request().method() === 'PUT');
      const p2 = page.waitForResponse(resp => resp.url().includes('service-3') && resp.request().method() === 'PUT');

      // Click Disable
      await page.getByRole('button', { name: 'Disable' }).click();

      await Promise.all([p1, p2]);
  });

    test('should delete services', async ({ page }) => {
      // Handle confirm dialog
      page.on('dialog', dialog => dialog.accept());

      await page.goto('/upstream-services');
      await expect(page.getByText('service-1')).toBeVisible();

      // Select service-2
      await page.getByRole('checkbox', { name: 'Select service-2' }).check();

      // Setup response wait
      const p1 = page.waitForResponse(resp => resp.url().includes('service-2') && resp.request().method() === 'DELETE');

      // Click Delete
      await page.getByRole('button', { name: 'Delete' }).click();

      await p1;

      // Verify it's gone
      await expect(page.getByText('service-2')).not.toBeVisible();
  });

});
