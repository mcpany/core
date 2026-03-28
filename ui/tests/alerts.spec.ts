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

    // Wait for the stats to load and MTTR should be 0s initially (no resolved alerts)
    await expect(page.getByText('0s').first()).toBeVisible();

    // The backend mock data has Active Critical: 1, Active Warning: 1
    // We expect these to be populated
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

    // After resolving an alert, MTTR should be > 0s since timestamp of generation is in the past
    await expect(page.getByText('0s').first()).toBeHidden({ timeout: 5000 });
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
