/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Advanced Logs Filtering', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to logs page
    await page.goto('/logs');
    // Wait for "Live Logs" to be visible
    await expect(page.getByRole('heading', { name: 'Live Logs' })).toBeVisible();
  });

  // Note: We cannot fully test "Time Range" exclusion with real logs because we cannot
  // inject historical logs into the backend without mocking.
  // We verified the logic via mock in development, but for E2E we verify the controls exist and don't crash.
  test('should have functioning time range controls', async ({ page }) => {
    const timeRangeTrigger = page.getByRole('combobox').nth(0);
    await expect(timeRangeTrigger).toBeVisible();
    await timeRangeTrigger.click();

    // Check if options are visible
    await expect(page.getByRole('option', { name: 'Last 1 Hour' })).toBeVisible();
    await expect(page.getByRole('option', { name: 'Last 24 Hours' })).toBeVisible();

    // Select an option
    await page.getByRole('option', { name: 'Last 24 Hours' }).click();

    // Verify selection applied (text content of trigger updates or dropdown closes)
    await expect(page.getByRole('option', { name: 'Last 24 Hours' })).toBeHidden(); // Dropdown closed

    // Ensure page didn't crash and controls are still there
    await expect(page.getByRole('heading', { name: 'Live Logs' })).toBeVisible();
  });

  test('should have functioning regex mode', async ({ page }) => {
    const searchInput = page.getByPlaceholder(/Search logs/);
    await expect(searchInput).toBeVisible();

    // Toggle Regex Mode
    const regexToggle = page.getByTitle('Toggle Regex Mode');
    await regexToggle.click();

    // Placeholder should update
    await expect(page.getByPlaceholder('Search logs (Regex)...')).toBeVisible();

    // Enter a regex
    await searchInput.fill('(INFO|WARN|ERROR)');

    // Verify it doesn't crash
    await expect(page.getByRole('heading', { name: 'Live Logs' })).toBeVisible();

    // Toggle back
    await regexToggle.click();
    await expect(page.getByPlaceholder('Search logs...')).toBeVisible();
  });
});
