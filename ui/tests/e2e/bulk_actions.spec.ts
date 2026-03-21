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

  test.beforeEach(async ({ request }) => {
    // Seed real services to backend
    for (const svc of testServices) {
      await request.delete(`/api/v1/services/${svc.name}`, { headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' } }).catch(() => {});

      const res = await request.post('/api/v1/services', {
        data: svc,
        headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' }
      });
      if (!res.ok()) {
          console.error(`Failed to seed service: ${await res.text()}`);
      }
      expect(res.ok()).toBeTruthy();
    }
  });

  test.afterEach(async ({ request }) => {
    // Clean up
    for (const svc of testServices) {
      await request.delete(`/api/v1/services/${svc.name}`, { headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' } }).catch(() => {});
    }
  });

  test('should select all services and show bulk actions', async ({ page }) => {
    await page.goto('/upstream-services');

    // Wait for services to load
    await expect(page.getByText('service-1-bulk')).toBeVisible();

    // Filter by tag so we only operate on our seeded test services to avoid interference
    await page.getByPlaceholder('Filter by tag...').fill('prod');
    // service-1 and service-3 have prod tag
    await expect(page.getByText('service-1-bulk')).toBeVisible();
    await expect(page.getByText('service-3-bulk')).toBeVisible();
    await expect(page.getByText('service-2-bulk')).not.toBeVisible();

    // Check "Select All" checkbox using role
    const selectAllCheckbox = page.getByRole('checkbox', { name: 'Select all' });
    await selectAllCheckbox.check();

    // Verify bulk action buttons appear (should be 2 because we filtered)
    await expect(page.getByText('2 selected')).toBeVisible();
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

  test('should toggle services', async ({ page, request }) => {
      await page.goto('/upstream-services');
      await expect(page.getByText('service-1-bulk')).toBeVisible();

      // Select service-1 and service-3
      await page.getByRole('checkbox', { name: 'Select service-1-bulk' }).check();
      await page.getByRole('checkbox', { name: 'Select service-3-bulk' }).check();

      // Click Disable
      await page.getByRole('button', { name: 'Disable' }).click();

      // Wait for the UI toast
      await expect(page.getByText('Services Disabled')).toBeVisible();

      // Allow a brief moment for state sync to database before asserting
      await page.waitForTimeout(500);

      const s1 = await request.get('/api/v1/services/service-1-bulk', { headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' } });
      const s1Data = await s1.json();
      expect(s1Data.service.disable).toBeTruthy();

      const s3 = await request.get('/api/v1/services/service-3-bulk', { headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' } });
      const s3Data = await s3.json();
      expect(s3Data.service.disable).toBeTruthy();
  });

  test('should delete services via bulk action', async ({ page, request }) => {
      await page.goto('/upstream-services');
      await expect(page.getByText('service-1-bulk')).toBeVisible();

      // Filter by tag so we only operate on our seeded test services to avoid interference
      await page.getByPlaceholder('Filter by tag...').fill('dev');
      // service-2-bulk has dev tag
      await expect(page.getByText('service-2-bulk')).toBeVisible();
      await expect(page.getByText('service-1-bulk')).not.toBeVisible();

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

      // Verify request in backend
      const res = await request.get('/api/v1/services/service-2-bulk', { headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' } });
      expect(res.status()).toBe(404);
  });

});
