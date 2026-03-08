/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Alerts Page', () => {
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

    // Check that stats values are fetched from the API and rendered
    // The backend mock data has Active Critical: 1, Active Warning: 1, MTTR: 14m, Total: >0 (we can just verify it's rendered)
    await expect(page.getByText('14m')).toBeVisible();

    // Check table content (mock data)
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

  test('should support bulk actions for alerts', async ({ page }) => {
    await page.goto('/alerts');

    // Wait for alerts to load
    await expect(page.getByText('High CPU Usage')).toBeVisible();
    await expect(page.getByText('High Memory Usage')).toBeVisible();

    // Get rows for the two alerts we want to bulk update
    const row1 = page.getByRole('row').filter({ hasText: 'High CPU Usage' });
    const row2 = page.getByRole('row').filter({ hasText: 'High Memory Usage' });

    // Select the checkboxes
    await row1.getByRole('checkbox').check();
    await row2.getByRole('checkbox').check();

    // Verify the bulk action toolbar appears
    await expect(page.getByText('2 alerts selected')).toBeVisible();

    // Click Bulk Acknowledge
    await page.getByRole('button', { name: 'Acknowledge' }).click();

    // Verify both rows now say "acknowledged"
    await expect(row1.getByText('acknowledged')).toBeVisible();
    await expect(row2.getByText('acknowledged')).toBeVisible();

    // The toolbar should disappear after action
    await expect(page.getByText('2 alerts selected')).toBeHidden();

    // Re-select them to test Bulk Resolve
    await row1.getByRole('checkbox').check();
    await row2.getByRole('checkbox').check();

    await expect(page.getByText('2 alerts selected')).toBeVisible();

    // Click Bulk Resolve
    await page.getByRole('button', { name: 'Resolve' }).click();

    // Verify both rows now say "resolved"
    await expect(row1.getByText('resolved')).toBeVisible();
    await expect(row2.getByText('resolved')).toBeVisible();

    // Test the "Select All" checkbox
    await page.getByRole('checkbox', { name: 'Select all alerts' }).check();
    // Assuming we have at least 5 alerts seeded
    await expect(page.getByText(/alerts selected/)).toBeVisible();

    // Uncheck "Select All"
    await page.getByRole('checkbox', { name: 'Select all alerts' }).uncheck();
    await expect(page.getByText(/alerts selected/)).toBeHidden();
  });
});
