/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Bulk Service Actions', () => {

  // Default API Key from configuration instead of process.env to prevent linting errors
  const HEADERS = { 'X-API-Key': 'test-token', 'Content-Type': 'application/json' };

  test.beforeEach(async ({ request }) => {
    const seedServices = [
      { name: "service-1", disable: false, tags: ["prod"], http_service: { address: "http://localhost:8001" } },
      { name: "service-2", disable: true, tags: ["dev"], http_service: { address: "http://localhost:8002" } },
      { name: "service-3", disable: false, tags: ["prod"], http_service: { address: "http://localhost:8003" } }
    ];

    for (const s of seedServices) {
      await request.post(`/api/v1/services`, { data: s, headers: HEADERS });
      await request.put(`/api/v1/services/${s.name}`, { data: s, headers: HEADERS });
      await request.put(`/api/v1/services/${s.name}/status`, { data: { disable: s.disable }, headers: HEADERS });
    }
  });

  test.afterEach(async ({ request }) => {
    const listRes = await request.get(`/api/v1/services`, { headers: HEADERS });
    if (listRes.ok()) {
      const { services } = await listRes.json();
      for (const service of (services || [])) {
        await request.delete(`/api/v1/services/${service.name}`, { headers: HEADERS });
      }
    }
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

  test('should toggle services', async ({ page, request }) => {
      await page.goto('/upstream-services');
      await expect(page.getByText('service-1')).toBeVisible();

      // Select service-1 and service-3
      await page.getByRole('checkbox', { name: 'Select service-1' }).check();
      await page.getByRole('checkbox', { name: 'Select service-3' }).check();

      // Click Disable
      await page.getByRole('button', { name: 'Disable' }).click();

      // Verify via actual backend state
      await expect.poll(async () => {
          const check1 = await request.get(`/api/v1/services/service-1`, { headers: HEADERS });
          const data1 = await check1.json();
          const check3 = await request.get(`/api/v1/services/service-3`, { headers: HEADERS });
          const data3 = await check3.json();
          return data1.disable === true && data3.disable === true;
      }, { timeout: 10000 }).toBeTruthy();
  });

    test('should delete services', async ({ page, request }) => {
      // Handle confirm dialog
      page.on('dialog', dialog => dialog.accept());

      await page.goto('/upstream-services');
      await expect(page.getByText('service-1')).toBeVisible();

      // Select service-2
      await page.getByRole('checkbox', { name: 'Select service-2' }).check();

      // Click Delete
      await page.getByRole('button', { name: 'Delete' }).click();

      // Verify via backend that it's gone
      await expect.poll(async () => {
          const res = await request.get(`/api/v1/services/service-2`, { headers: HEADERS });
          return res.status(); // Should be 404
      }, { timeout: 10000 }).toBe(404);
  });

});
