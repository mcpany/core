/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Bulk Service Actions', () => {

  test.beforeEach(async ({ page }) => {
    // Mock services API
    await page.route('**/api/v1/services', async route => {
        await route.fulfill({
            json: [
                { name: "service-1", httpService: { address: "http://localhost:8001" }, disable: false, tags: ["prod"] },
                { name: "service-2", httpService: { address: "http://localhost:8002" }, disable: true, tags: ["dev"] },
                { name: "service-3", httpService: { address: "http://localhost:8003" }, disable: false, tags: ["prod"] }
            ]
        });
    });

     // Mock doctor API
    await page.route('**/doctor', async route => {
        await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ status: 'healthy', checks: {} })
        });
    });
  });

  test('should select all services and show bulk actions', async ({ page }) => {
    await page.goto('/services');

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
     await page.goto('/services');
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
      // Mock the toggle API
      const toggleRequests: string[] = [];
      await page.route('**/api/v1/services/*', async route => {
          if (route.request().method() === 'PUT') {
              toggleRequests.push(route.request().url());
              await route.fulfill({ status: 200, json: {} });
          } else {
              await route.continue();
          }
      });

      await page.goto('/services');
      await expect(page.getByText('service-1')).toBeVisible();

      // Select service-1 and service-3
      await page.getByRole('checkbox', { name: 'Select service-1' }).check();
      await page.getByRole('checkbox', { name: 'Select service-3' }).check();

      // Click Disable
      await page.getByRole('button', { name: 'Disable' }).click();

      // Verify requests
      await expect.poll(() => toggleRequests.length).toBe(2);
      expect(toggleRequests.some(url => url.includes('service-1'))).toBeTruthy();
      expect(toggleRequests.some(url => url.includes('service-3'))).toBeTruthy();
  });

    test('should delete services', async ({ page }) => {
      // Mock the delete API
      const deleteRequests: string[] = [];
      await page.route('**/api/v1/services/*', async route => {
          if (route.request().method() === 'DELETE') {
            deleteRequests.push(route.request().url());
            await route.fulfill({ status: 200 });
          } else {
            await route.continue();
          }
      });

      // Handle confirm dialog
      page.on('dialog', dialog => dialog.accept());

      await page.goto('/services');
      await expect(page.getByText('service-1')).toBeVisible();

      // Select service-2
      await page.getByRole('checkbox', { name: 'Select service-2' }).check();

      // Click Delete
      await page.getByRole('button', { name: 'Delete' }).click();

      // Wait a bit for async calls
      await expect.poll(() => deleteRequests.length).toBe(1);
      expect(deleteRequests[0]).toContain('service-2');
  });

});
