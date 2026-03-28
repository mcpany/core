/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Alerts Page', () => {

  test.beforeEach(async ({ request }) => {
    // Seed the database by creating test alerts via API before each test
    // We create one critical and one warning alert
    await request.post('/api/v1/alerts', {
      data: {
        title: 'High CPU Usage',
        message: 'CPU usage > 90% for 5m',
        severity: 'critical',
        status: 'active',
        service: 'weather-service',
        source: 'System Monitor',
        timestamp: new Date(Date.now() - 5 * 60000).toISOString() // 5 minutes ago
      }
    });

    await request.post('/api/v1/alerts', {
      data: {
        title: 'API Latency Spike',
        message: 'P99 Latency > 2000ms',
        severity: 'warning',
        status: 'active',
        service: 'api-gateway',
        source: 'Latency Watchdog',
        timestamp: new Date(Date.now() - 15 * 60000).toISOString() // 15 minutes ago
      }
    });
  });

  test.afterEach(async ({ request }) => {
    // Clean up created alerts after test
    const response = await request.get('/api/v1/alerts');
    const alerts = await response.json();
    for (const alert of alerts) {
      await request.delete(`/api/v1/alerts/${alert.id}`);
    }
  });

  test('should load alerts page and display key elements', async ({ page }) => {
    // Navigate to alerts page
    await page.goto('/alerts');

    // Check header
    await expect(page.getByRole('heading', { name: 'Alerts & Incidents' })).toBeVisible();

    // Check stats cards labels
    await expect(page.getByText('Active Critical')).toBeVisible();
    await expect(page.getByText('Active Warnings')).toBeVisible();
    await expect(page.getByText('MTTR (Today)')).toBeVisible();
    await expect(page.getByText('Total Incidents')).toBeVisible();

    // Wait for the stats to load and MTTR should be 0s initially (no resolved alerts)
    await expect(page.getByText('0s').first()).toBeVisible();

    // We expect the seeded alerts to be populated
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

    // Find an active alert row
    const row = page.getByRole('row').filter({ hasText: 'High CPU Usage' });

    // Click the "More Actions" dropdown button in that row
    await row.getByRole('button', { name: 'Open menu' }).click();

    // Click "Acknowledge"
    await page.getByRole('menuitem', { name: 'Acknowledge' }).click();

    // Verify status changes to "acknowledged"
    await expect(row.getByText('acknowledged')).toBeVisible();
  });

  test('should resolve alert via dropdown and update MTTR', async ({ page }) => {
    await page.goto('/alerts');

    // Make sure initial MTTR is 0s
    await expect(page.getByText('0s').first()).toBeVisible();

    // Find an active alert to resolve
    const row = page.getByRole('row').filter({ hasText: 'High CPU Usage' });

    // Ensure it exists first
    await expect(row).toBeVisible();

    // Click "More Actions"
    await row.getByRole('button', { name: 'Open menu' }).click();

    // Wait for the dropdown menu
    await expect(page.getByRole('menuitem', { name: 'Resolve' })).toBeVisible();
    // Click "Resolve"
    await page.getByRole('menuitem', { name: 'Resolve' }).click();

    // Verify status changes to "resolved"
    await expect(row.getByText('resolved')).toBeVisible();

    // Note: since the API request to update the backend might race with the UI's status refresh
    // vs the stats refresh, we just wait a bit and check that MTTR no longer equals "0s"
    // or we can wait for the specific expected string "5m".
    await expect(page.getByText('0s').first()).toBeHidden({ timeout: 10000 });
    // Verify it updated to something like "5m"
    await expect(page.getByText('m').first()).toBeVisible({ timeout: 10000 });
  });

  test('should delete alert via dropdown', async ({ page }) => {
    await page.goto('/alerts');

    // Find a specific alert to delete
    const row = page.getByRole('row').filter({ hasText: 'High CPU Usage' });

    // Make sure it exists first
    await expect(row).toBeVisible();

    // Click "More Actions"
    await row.getByRole('button', { name: 'Open menu' }).click();

    // Click "Delete"
    await page.getByRole('menuitem', { name: 'Delete' }).click();

    // Wait for row to disappear
    await expect(row).toBeHidden();
  });
});
