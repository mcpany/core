/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Alerts Page', () => {
  test.beforeEach(async ({ page }) => {
    // Consolidated Mock for /api/v1/alerts and /api/v1/alerts/*
    // Using regex to match both list and detail/update URLs reliably
    await page.route(/\/api\/v1\/alerts/, async route => {
      const method = route.request().method();
      const url = route.request().url();
      console.log(`Intercepted ${method} ${url}`);

      // Handle List (GET /api/v1/alerts)
      if (method === 'GET' && !url.match(/\/api\/v1\/alerts\/.+/)) {
        const json = [
          { id: '1', title: 'High CPU Usage', message: 'CPU > 90%', severity: 'critical', status: 'active', service: 'weather-service', timestamp: new Date().toISOString() },
          { id: '2', title: 'API Latency Spike', message: 'Latency > 2s', severity: 'warning', status: 'active', service: 'api-gateway', timestamp: new Date().toISOString() },
          { id: '3', title: 'Disk Space Low', message: 'Volume /data at 90%', severity: 'warning', status: 'acknowledged', service: 'database', timestamp: new Date().toISOString() }
        ];
        await route.fulfill({ json });
        return;
      }

      // Handle Update (PATCH /api/v1/alerts/{id})
      if (method === 'PATCH') {
        const body = route.request().postDataJSON();
        const id = url.split('/').pop();

        // Preserve original fields to keep row visible
        let original: any = {};
        if (id === '1') original = { title: 'High CPU Usage', severity: 'critical', service: 'weather-service' };
        else if (id === '2') original = { title: 'API Latency Spike', severity: 'warning', service: 'api-gateway' };
        else if (id === '3') original = { title: 'Disk Space Low', severity: 'warning', service: 'database' };

        await route.fulfill({
          json: {
            id,
            status: body.status,
            title: original.title || 'Mock Alert',
            message: 'Mock Message',
            severity: original.severity || 'info',
            service: original.service || 'mock-service',
            timestamp: new Date().toISOString()
          }
        });
        return;
      }

      await route.continue();
    });
  });

  test('should load alerts page and display key elements', async ({ page }) => {
    // Navigate to alerts page
    await page.goto('/alerts');

    // Check header
    await expect(page.getByRole('heading', { name: 'Alerts & Incidents' })).toBeVisible();

    // Check stats cards - Use exact match because description might contain partial text
    await expect(page.getByText('Active Critical', { exact: true })).toBeVisible();
    await expect(page.getByText('MTTR (Today)', { exact: true })).toBeVisible();

    // Check table content (mocked data)
    await expect(page.getByText('High CPU Usage')).toBeVisible();
    await expect(page.getByText('API Latency Spike')).toBeVisible();
  });

  test('should filter alerts', async ({ page }) => {
    await page.goto('/alerts');

    // Type in search box - use getByPlaceholder if available, else locator
    const searchBox = page.locator('input[placeholder="Search alerts by title, message, service..."]');
    await searchBox.fill('CPU');

    // Should see CPU alert
    await expect(page.getByText('High CPU Usage')).toBeVisible();

    // Should NOT see Latency alert
    await expect(page.getByText('API Latency Spike')).toBeHidden();
  });

  test('should open create rule dialog', async ({ page }) => {
    await page.goto('/alerts');

    // Click "New Alert Rule" button
    await page.getByRole('button', { name: 'New Alert Rule' }).click();

    // Check dialog opens
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Create Alert Rule' })).toBeVisible();

    // Close it
    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page.getByRole('dialog')).toBeHidden();
  });
  test('should acknowledge alert via dropdown', async ({ page }) => {
    await page.goto('/alerts');

    // Find an active alert row (mock data usually has some)
    // We target the row with "High CPU Usage" which is active in mock
    const row = page.getByRole('row').filter({ hasText: 'High CPU Usage' });

    // Click the "More Actions" dropdown button in that row
    await row.getByRole('button', { name: 'Open menu' }).click();

    // Click "Acknowledge"
    await page.getByRole('menuitem', { name: 'Acknowledge' }).click();

    // Verify status changes to "acknowledged"
    await expect(row.getByText('acknowledged')).toBeVisible();
  });

  test('should resolve alert via dropdown', async ({ page }) => {
    await page.goto('/alerts');

    // Find an acknowledged or active alert
    const row = page.getByRole('row').filter({ hasText: 'Disk Space Low' });

    // Click "More Actions"
    await row.getByRole('button', { name: 'Open menu' }).click();

    // Click "Resolve"
    await page.getByRole('menuitem', { name: 'Resolve' }).click();

    // Verify status changes to "resolved"
    await expect(row.getByText('resolved')).toBeVisible();
  });
});
