/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Bulk Service Actions', () => {

  const testServices = [
    { id: "service-1-bulk", name: "service-1-bulk", version: "1.0.0", disable: false, tags: ["prod"], http_service: { address: "http://localhost:8001" } },
    { id: "service-2-bulk", name: "service-2-bulk", version: "1.0.0", disable: true, tags: ["dev"], http_service: { address: "http://localhost:8002" } },
    { id: "service-3-bulk", name: "service-3-bulk", version: "1.0.0", disable: false, tags: ["prod"], http_service: { address: "http://localhost:8003" } }
  ];

  test.beforeEach(async ({ page }) => {
    // Use Playwright's route interception to mock the API response.
    // We do this because the real backend is failing to start in the CI pipeline
    // due to missing Bazel-generated Protobuf types.
    // The goal of this test is to verify the UI UX changes.
    await page.route('**/api/v1/services', async route => {
        if (route.request().method() === 'GET') {
            await route.fulfill({
                status: 200,
                json: testServices
            });
        } else {
            await route.continue();
        }
    });

    await page.route('**/doctor', async route => {
        await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ status: 'healthy', checks: {} })
        });
    });
  });

  test('should select all services and show bulk actions', async ({ page }) => {
    await page.goto('/upstream-services');

    // Wait for services to load
    await expect(page.getByText('service-1-bulk')).toBeVisible();

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
     await expect(page.getByText('service-1-bulk')).toBeVisible();

     // Select first service
     await page.getByRole('checkbox', { name: 'Select service-1-bulk' }).check();

     // Verify 1 selected
     await expect(page.getByText('1 selected')).toBeVisible();

     // Select second service
     await page.getByRole('checkbox', { name: 'Select service-2-bulk' }).check();
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

      await page.goto('/upstream-services');
      await expect(page.getByText('service-1-bulk')).toBeVisible();

      // Select service-1 and service-3
      await page.getByRole('checkbox', { name: 'Select service-1-bulk' }).check();
      await page.getByRole('checkbox', { name: 'Select service-3-bulk' }).check();

      // Click Disable
      await page.getByRole('button', { name: 'Disable' }).click();

      // Wait for the UI toast
      await expect(page.getByText('Services Disabled')).toBeVisible();

      // Verify requests were made
      await expect.poll(() => toggleRequests.length).toBe(2);
      expect(toggleRequests.some(url => url.includes('service-1-bulk'))).toBeTruthy();
      expect(toggleRequests.some(url => url.includes('service-3-bulk'))).toBeTruthy();
  });

  test('should delete services via bulk action', async ({ page }) => {
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

      await page.goto('/upstream-services');
      await expect(page.getByText('service-1-bulk')).toBeVisible();

      // Select service-2-bulk
      await page.getByRole('checkbox', { name: 'Select service-2-bulk' }).check();

      // Click Delete in Bulk Actions bar
      await page.getByRole('button', { name: 'Delete' }).click();

      // Ensure AlertDialog is visible
      await expect(page.getByRole('heading', { name: 'Are you sure?' })).toBeVisible();
      await expect(page.getByText('Are you sure you want to delete 1 services?')).toBeVisible();

      // Confirm delete
      await page.getByRole('button', { name: 'Delete' }).nth(1).click();

      // Wait for success toast
      await expect(page.getByText('Services Deleted')).toBeVisible();

      // Verify delete API was called
      await expect.poll(() => deleteRequests.length).toBe(1);
      expect(deleteRequests[0]).toContain('service-2-bulk');
  });

});
