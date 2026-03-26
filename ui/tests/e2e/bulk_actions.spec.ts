/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Bulk Service Actions', () => {

  test.beforeEach(async ({ page, request }) => {
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
  });

  test('should select all services and show bulk actions', async ({ page }) => {
    await page.goto('/upstream-services');

    // Wait for services to load
    await expect(page.getByText('Payment Gateway')).toBeVisible();
    await expect(page.getByText('User Service')).toBeVisible();

    // Check "Select All" checkbox using role
    const selectAllCheckbox = page.getByRole('checkbox', { name: 'Select all' });
    await selectAllCheckbox.check();

    // Verify bulk action buttons appear
    await expect(page.getByText('selected')).toBeVisible(); // 3 or more selected
    await expect(page.getByRole('button', { name: 'Enable' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Disable' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Delete' })).toBeVisible();
  });

  test('should select individual services', async ({ page }) => {
     await page.goto('/upstream-services');
     await expect(page.getByText('Payment Gateway')).toBeVisible();

     // Select first service
     await page.getByRole('checkbox', { name: 'Select Payment Gateway' }).check();

     // Verify 1 selected
     await expect(page.getByText('1 selected')).toBeVisible();

     // Select second service
     await page.getByRole('checkbox', { name: 'Select User Service' }).check();
     await expect(page.getByText('2 selected')).toBeVisible();
  });

  test('should toggle services', async ({ page, request }) => {
      await page.goto('/upstream-services');
      await expect(page.getByText('Payment Gateway')).toBeVisible();

      // Select Payment Gateway and User Service
      await page.getByRole('checkbox', { name: 'Select Payment Gateway' }).check();
      await page.getByRole('checkbox', { name: 'Select User Service' }).check();

      // Wait a moment for state to be fully settled (debounce)
      await page.waitForTimeout(500);

      // Click Disable
      await page.getByRole('button', { name: 'Disable' }).click();
      // Use polling to verify API updates as waitForResponse might miss rapidly dispatched requests in bulk

      // Verify via API that they are disabled
      // The E2E tests mock the API locally so it doesn't really update the API if not properly configured. We'll check the UI toggle.
      // After disabling, the button should turn to Enable or the Toast should be visible.
      // Let's use UI assertion since we know it clicked Disable and optimistic UI applied it.
      await page.waitForTimeout(500);

      await page.getByRole('checkbox', { name: 'Select Payment Gateway' }).check();
      await page.getByRole('checkbox', { name: 'Select User Service' }).check();

      await expect(page.getByRole('button', { name: 'Enable' })).toBeVisible({ timeout: 5000 });

      // Verify via API that they are disabled
      await expect.poll(async () => {
          const res1 = await request.get(`/api/v1/services/${encodeURIComponent('Payment Gateway')}`, {
              headers: { 'X-API-Key': 'test-token' }
          });
          if (!res1.ok()) return false;
          const data1 = await res1.json();
          // Depending on API response format it's either data.service.disable or data.disable
          return data1.service?.disable === true || data1.disable === true;
      }, { timeout: 10000 }).toBeTruthy();

      await expect.poll(async () => {
          const res3 = await request.get(`/api/v1/services/${encodeURIComponent('User Service')}`, {
              headers: { 'X-API-Key': 'test-token' }
          });
          if (!res3.ok()) return false;
          const data3 = await res3.json();
          return data3.service?.disable === true || data3.disable === true;
      }, { timeout: 10000 }).toBeTruthy();
  });

  test('should delete services', async ({ page, request }) => {
      // Handle confirm dialog
      page.on('dialog', dialog => dialog.accept());

      await page.goto('/upstream-services');
      await expect(page.getByText('User Service')).toBeVisible();

      // Select User Service
      await page.getByRole('checkbox', { name: 'Select User Service' }).check();

      // Click Delete
      await page.getByRole('button', { name: 'Delete' }).click();

      // Wait for success toast or disappearance
      await expect(page.getByText('User Service')).toBeHidden();

      // Verify via API that it's deleted
      const res2 = await request.get('/api/v1/services/User Service');
      expect(res2.status()).toBe(404);
  });

});
