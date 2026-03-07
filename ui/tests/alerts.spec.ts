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

  test('should bulk acknowledge alerts', async ({ page }) => {
    await page.goto('/alerts');

    // Wait for the alerts table to be populated
    await expect(page.getByText('High CPU Usage')).toBeVisible();
    await expect(page.getByText('API Latency Spike')).toBeVisible();

    // Select the first alert row using the checkbox
    const firstRow = page.getByRole('row').filter({ hasText: 'High CPU Usage' });
    await firstRow.getByRole('checkbox').click();

    // Select the second alert row using the checkbox
    const secondRow = page.getByRole('row').filter({ hasText: 'API Latency Spike' });
    await secondRow.getByRole('checkbox').click();

    // Verify that the bulk action bar is visible
    const actionBarText = page.getByText('2 selected');
    await expect(actionBarText).toBeVisible();

    // Click the bulk "Acknowledge" button
    const acknowledgeButton = page.getByRole('button', { name: 'Acknowledge', exact: true });
    await acknowledgeButton.click();

    // Verify toast notification appears
    await expect(page.getByText('Bulk Update Successful')).toBeVisible();

    // Verify that the selected alerts' statuses changed to "acknowledged"
    await expect(firstRow.getByText('acknowledged')).toBeVisible();
    await expect(secondRow.getByText('acknowledged')).toBeVisible();

    // Verify the bulk action bar is hidden after success
    await expect(actionBarText).toBeHidden();
  });
});
