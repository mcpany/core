/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Settings Page', () => {
  test('should successfully clear cache', async ({ page }) => {
    // Navigate to settings page
    await page.goto('/settings');

    // Wait for the page to load
    await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible();

    // Verify the Global Config tab is selected by default
    await expect(page.getByRole('tab', { name: 'Global Config' })).toHaveAttribute('aria-selected', 'true');

    // Ensure the System Actions section is visible
    await expect(page.getByRole('heading', { name: 'System Actions' })).toBeVisible();

    // Setup request interception to verify the API call
    const clearCacheRequestPromise = page.waitForRequest(request =>
      request.url().includes('/api/v1/admin/cache/clear') && request.method() === 'POST'
    );
    const clearCacheResponsePromise = page.waitForResponse(response =>
      response.url().includes('/api/v1/admin/cache/clear') && response.status() === 200
    );

    // Find and click the Clear Cache button
    const clearCacheButton = page.getByRole('button', { name: 'Clear Cache' });
    await expect(clearCacheButton).toBeVisible();
    await clearCacheButton.click();

    // Wait for the network request and response to complete
    const request = await clearCacheRequestPromise;
    expect(request.postDataJSON()).toEqual({});

    const response = await clearCacheResponsePromise;
    expect(response.ok()).toBeTruthy();

    // Verify the success toast appears
    await expect(page.getByText('Cache Cleared')).toBeVisible();
    await expect(page.getByText('The system cache has been successfully cleared.')).toBeVisible();

    // Verify button text reverts
    await expect(page.getByRole('button', { name: 'Clear Cache' })).toBeVisible();
  });
});
