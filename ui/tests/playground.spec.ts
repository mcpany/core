/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Tool Configuration', () => {
  // Use real backend data (get_weather tool from config.minimal.yaml)
  test('should allow configuring and running a tool via unified runner', async ({ page }) => {
    await page.goto('/playground');

    // Wait for sidebar to load tools (real data)
    await expect(page.getByText('get_weather')).toBeVisible({ timeout: 30000 });

    // Click tool card for get_weather in Sidebar
    // This should switch the main pane to "Tool Runner" mode
    await page.locator('.group').filter({ hasText: 'get_weather' }).click();

    // Verify "Tool Runner" tab is active
    await expect(page.getByRole('tab', { name: 'Tool Runner' })).toHaveAttribute('data-state', 'active');

    // Verify Tool Runner content
    // Should see "get_weather" in the header
    await expect(page.getByRole('heading', { name: 'get_weather' })).toBeVisible();

    // Switch to JSON tab to input arguments (since schema might be empty in minimal config)
    await page.getByRole('tab', { name: 'JSON' }).first().click();

    // Fill JSON arguments
    await page.getByPlaceholder('{}').fill('{"city": "San Francisco"}');

    // Run Tool (Execute button in the runner header)
    await page.getByRole('button', { name: 'Execute' }).click();

    // Verify Result
    // Expect "Success" badge
    await expect(page.getByText('Success')).toBeVisible({ timeout: 10000 });

    // Verify "Result" section is visible
    await expect(page.getByText('Result', { exact: true })).toBeVisible();
  });
});
