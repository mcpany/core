/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

const MOCK_ALERTS = [
  {
    id: '1',
    title: 'High CPU Usage',
    message: 'CPU usage > 90%',
    severity: 'critical',
    status: 'active',
    service: 'backend-api',
    timestamp: new Date().toISOString(),
    source: 'system'
  },
  {
    id: '2',
    title: 'API Latency Spike',
    message: 'p99 > 500ms',
    severity: 'warning',
    status: 'active',
    service: 'auth-service',
    timestamp: new Date().toISOString(),
    source: 'datadog'
  },
  {
    id: '3',
    title: 'Disk Space Low',
    message: 'Free space < 10%',
    severity: 'info',
    status: 'active',
    service: 'database',
    timestamp: new Date().toISOString(),
    source: 'aws'
  }
];

test.describe('Alerts Page', () => {
  test.beforeEach(async ({ page }) => {
    // Mock the Alerts API
    await page.route('*/**/api/v1/alerts', async route => {
      const method = route.request().method();
      if (method === 'GET') {
        await route.fulfill({ json: MOCK_ALERTS });
      } else {
        await route.continue();
      }
    });

    // Mock Status Updates
    await page.route('*/**/api/v1/alerts/*', async route => {
        const method = route.request().method();
        if (method === 'PATCH') {
            const url = route.request().url();
            const id = url.split('/').pop();
            const body = route.request().postDataJSON();

            const alert = MOCK_ALERTS.find(a => a.id === id);
            if (alert) {
                 await route.fulfill({ json: { ...alert, status: body.status } });
            } else {
                 await route.fulfill({ status: 404 });
            }
        } else {
            await route.continue();
        }
    });
  });

  test('should load alerts page and display key elements', async ({ page }) => {
    // Navigate to alerts page
    await page.goto('/alerts');

    // Check header
    await expect(page.getByRole('heading', { name: 'Alerts & Incidents' })).toBeVisible();

    // Check stats cards
    // Use strict matching or specific container if needed, but text should exist
    // "Active Critical" is in the card title
    await expect(page.locator('.text-sm.font-medium', { hasText: 'Active Critical' })).toBeVisible();

    // NOTE: "MTTR (Today)" is NOT in my implementation of alert-stats.tsx
    // I implemented: Active Critical, Active Warnings, Resolved (Today), Total Incidents (Today)
    // So I should check for those.
    await expect(page.locator('.text-sm.font-medium', { hasText: 'Active Warnings' })).toBeVisible();
    await expect(page.locator('.text-sm.font-medium', { hasText: 'Total Incidents (Today)' })).toBeVisible();

    // Check table content (mock data)
    await expect(page.getByText('High CPU Usage')).toBeVisible();
    await expect(page.getByText('API Latency Spike')).toBeVisible();
  });

  test('should filter alerts', async ({ page }) => {
    await page.goto('/alerts');

    // Type in search box - use getByPlaceholder if available, else locator
    const searchBox = page.getByPlaceholder('Search alerts by title, message, service...');
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

    // Verify status changes to "acknowledged" (Visual check only since we mock response but component state updates)
    // In AlertList, handleStatusChange updates local state optimistically or after response.
    // Our mock returns the updated object.
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
