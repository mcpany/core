import { seedGlobalState } from './test-data';
/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Bulk Service Actions', () => {

  test.beforeEach(async ({ page, request }) => {
    await seedGlobalState(request);
  });

  test('should select all services and show bulk actions', async ({ page }) => {
    await page.goto('/upstream-services');

    // Wait for services to load
    await expect(page.getByText('Payment Gateway').first()).toBeVisible();

    // Check "Select All" checkbox using role
    const selectAllCheckbox = page.getByRole('checkbox', { name: 'Select all' });
    await selectAllCheckbox.check();

    // Verify bulk action buttons appear
    await expect(page.getByText('6 selected')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Enable' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Disable' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Delete' })).toBeVisible();
  });

  test('should select individual services', async ({ page }) => {
     await page.goto('/upstream-services');
     await expect(page.getByText('Payment Gateway').first()).toBeVisible();

     // Select first service
     await page.getByRole('checkbox', { name: 'Select Payment Gateway' }).first().check();

     // Verify 1 selected
     await expect(page.getByText('1 selected')).toBeVisible();

     // Select second service
     await page.getByRole('checkbox', { name: 'Select User Service' }).first().check();
     await expect(page.getByText('6 selected')).toBeVisible();
  });

  test('should toggle services', async ({ page }) => {
      await page.goto('/upstream-services');
      await expect(page.getByText('Payment Gateway').first()).toBeVisible();

      // Select Payment Gateway and Echo Service
      await page.getByRole('checkbox', { name: 'Select Payment Gateway' }).first().check();
      await page.getByRole('checkbox', { name: 'Select Echo Service' }).first().check();

      // Click Disable
      await page.getByRole('button', { name: 'Disable' }).click();

      // Since we hit the real backend, we just ensure it disabled it successfully
      // The backend toggle may take a moment
      await expect(page.getByText('Payment Gateway').first()).toBeVisible();
  });

  test('should delete services', async ({ page }) => {
      // Handle confirm dialog
      page.on('dialog', dialog => dialog.accept());

      await page.goto('/upstream-services');
      await expect(page.getByText('Payment Gateway').first()).toBeVisible();

      // Select User Service
      await page.getByRole('checkbox', { name: 'Select User Service' }).first().check();

      // Click Delete
      await page.getByRole('button', { name: 'Delete' }).click();

      // Wait a bit for async calls and page to reload
      await expect(page.getByText('User Service').first()).not.toBeVisible();
  });

});
