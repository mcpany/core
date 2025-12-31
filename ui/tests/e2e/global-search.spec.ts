/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Global Search', () => {
  test('should open and display dynamic content via keyboard shortcut', async ({ page }) => {
    // Navigate to the dashboard
    await page.goto('/');

    // Wait for the page to load using a more specific selector
    await expect(page.locator('h2:has-text("Dashboard")')).toBeVisible();

    // Simulate Cmd+K (Meta+K on Mac, Control+K on Windows/Linux)
    if (process.platform === 'darwin') {
      await page.keyboard.press('Meta+k');
    } else {
      await page.keyboard.press('Control+k');
    }

    // Verify the dialog is open
    await expect(page.locator('div[role="dialog"]')).toBeVisible();
    await expect(page.locator('input[placeholder="Type a command or search..."]')).toBeVisible();

    // Wait for data to load

    // Check for Suggestions
    await expect(page.getByText('Suggestions')).toBeVisible();

    // Check for dynamic content
    // Services
    await expect(page.getByText('weather-service').first()).toBeVisible();

    // Tools
    await expect(page.getByText('get_weather').first()).toBeVisible();

    // Resources
    await expect(page.getByText('notes.txt').first()).toBeVisible();

    // Prompts
    await expect(page.getByText('summarize_file').first()).toBeVisible();

    // Type in the search box to filter
    const input = page.locator('input[placeholder="Type a command or search..."]');
    await input.fill('weather');

    // Verify filtering works
    await expect(page.getByText('weather-service').first()).toBeVisible();
    await expect(page.getByText('get_weather').first()).toBeVisible();
    await expect(page.getByText('local-files')).not.toBeVisible();

    // Take a screenshot for audit
    await page.screenshot({ path: '.audit/ui/global_search_audit.png' });
  });

  test('should open command palette by clicking the search button', async ({ page }) => {
     await page.goto('/');

     // Find the button with text "Search" or similar
     // The button text changes based on screen size ("Search feature..." vs "Search...")
     const searchButton = page.locator('button').filter({ hasText: /Search/ }).first();
     await searchButton.click();

     const commandDialog = page.locator('[role="dialog"]');
     await expect(commandDialog).toBeVisible();
  });
});
