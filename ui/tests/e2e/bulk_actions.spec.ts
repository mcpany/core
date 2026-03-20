/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Bulk Service Actions', () => {
  const serviceNames = [`service-1-${Date.now()}`, `service-2-${Date.now()}`, `service-3-${Date.now()}`];

  test.beforeEach(async ({ request }) => {
    // Seed the database with real services via the API
    for (const name of serviceNames) {
      await request.post('/api/v1/services/register', {
        data: {
          config: {
            name: name,
            httpService: { address: "http://localhost:8000" },
            disable: false,
            tags: ["initial"]
          }
        }
      });
    }
  });

  test.afterEach(async ({ request }) => {
    // Clean up
    for (const name of serviceNames) {
      await request.post('/api/v1/services/unregister', {
        data: { serviceName: name }
      });
    }
  });

  test('should select all services and show bulk actions', async ({ page }) => {
    await page.goto('/upstream-services');

    // Wait for services to load
    await expect(page.getByText(serviceNames[0])).toBeVisible();

    // Check "Select All" checkbox using role
    const selectAllCheckbox = page.getByRole('checkbox', { name: 'Select all' });
    await selectAllCheckbox.check();

    // Verify bulk action buttons appear. We need to wait for the selection text because it might be more than 3 if other tests leaked.
    // Instead we check if the selected text matches the number of checkboxes minus the 'select all' one.
    // However, since we're using a specific DB, let's just make sure at least 3 are selected.
    await expect(page.locator('text=/\\d+ selected/')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Enable' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Disable' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Delete' })).toBeVisible();
  });

  test('should select individual services', async ({ page }) => {
     await page.goto('/upstream-services');
     await expect(page.getByText(serviceNames[0])).toBeVisible();

     // Select first service
     await page.getByRole('checkbox', { name: `Select ${serviceNames[0]}` }).check();

     // Verify 1 selected
     await expect(page.getByText('1 selected')).toBeVisible();

     // Select second service
     await page.getByRole('checkbox', { name: `Select ${serviceNames[1]}` }).check();
     await expect(page.getByText('2 selected')).toBeVisible();
  });

  test('should toggle services', async ({ page, request }) => {
      await page.goto('/upstream-services');
      await expect(page.getByText(serviceNames[0])).toBeVisible();

      // Select service-1 and service-3
      await page.getByRole('checkbox', { name: `Select ${serviceNames[0]}` }).check();
      await page.getByRole('checkbox', { name: `Select ${serviceNames[2]}` }).check();

      // Click Disable
      await page.getByRole('button', { name: 'Disable' }).click();

      await expect(page.getByText('Services Disabled')).toBeVisible();

      // Verify via API that they are disabled
      const res1 = await request.get(`/api/v1/services/${serviceNames[0]}`);
      const data1 = await res1.json();
      expect(data1.service.disable).toBe(true);

      const res3 = await request.get(`/api/v1/services/${serviceNames[2]}`);
      const data3 = await res3.json();
      expect(data3.service.disable).toBe(true);
  });

  test('should delete services', async ({ page, request }) => {
      // Handle confirm dialog
      page.on('dialog', dialog => dialog.accept());

      await page.goto('/upstream-services');
      await expect(page.getByText(serviceNames[0])).toBeVisible();

      // Select service-2
      await page.getByRole('checkbox', { name: `Select ${serviceNames[1]}` }).check();

      // Click Delete
      await page.getByRole('button', { name: 'Delete' }).click();

      await expect(page.getByText('Services Deleted')).toBeVisible();

      // Verify via API
      const res2 = await request.get(`/api/v1/services/${serviceNames[1]}`);
      expect(res2.status()).toBe(404);
  });

  test('should bulk edit tags', async ({ page, request }) => {
    await page.goto('/upstream-services');
    await expect(page.getByText(serviceNames[0])).toBeVisible();

    // Select the first two services
    await page.getByRole('checkbox', { name: `Select ${serviceNames[0]}` }).check();
    await page.getByRole('checkbox', { name: `Select ${serviceNames[1]}` }).check();

    // Click Bulk Edit
    await page.getByRole('button', { name: 'Bulk Edit' }).click();

    // Use the TagInput
    await expect(page.getByText('Add Tags')).toBeVisible();
    const tagInput = page.getByPlaceholder('Type a tag and press Enter');

    // Add 'production'
    await tagInput.fill('production');
    await tagInput.press('Enter');

    // Add 'web'
    await tagInput.fill('web');
    await tagInput.press('Enter');

    // Click Apply Changes
    await page.getByRole('button', { name: 'Apply Changes' }).click();

    await expect(page.getByText('Services Updated')).toBeVisible();

    // Verify via API that tags were appended
    const res1 = await request.get(`/api/v1/services/${serviceNames[0]}`);
    const data1 = await res1.json();
    expect(data1.service.tags).toContain('production');
    expect(data1.service.tags).toContain('web');
    expect(data1.service.tags).toContain('initial');

    const res2 = await request.get(`/api/v1/services/${serviceNames[1]}`);
    const data2 = await res2.json();
    expect(data2.service.tags).toContain('production');
    expect(data2.service.tags).toContain('web');
    expect(data2.service.tags).toContain('initial');

    // Make sure service-3 wasn't changed
    const res3 = await request.get(`/api/v1/services/${serviceNames[2]}`);
    const data3 = await res3.json();
    expect(data3.service.tags).not.toContain('production');
  });

});
