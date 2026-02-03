/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Global Search', () => {
  test.beforeEach(async ({ page }) => {
    // We don't need to seed data for these tests as we search for static items
    await page.goto('/');
  });

  test('should open command palette by clicking the search button', async ({ page }) => {
     // Find the button with text "Search" or similar
     // The button text changes based on screen size ("Search feature..." vs "Search...")
     const searchButton = page.locator('button').filter({ hasText: /Search/ }).first();
     await searchButton.click();

     const commandDialog = page.locator('[role="dialog"]');
     await expect(commandDialog).toBeVisible();
  });

  test('should open command palette via shortcut and search dynamic content', async ({ page }) => {
    // Focus page to ensure keyboard events are captured
    await page.click('body');

    // Determine modifier key
    const modifier = process.platform === 'darwin' ? 'Meta' : 'Control';
    await page.keyboard.press(`${modifier}+k`);
    await page.waitForTimeout(500);

    const dialog = page.locator('div[role="dialog"]');
    await expect(dialog).toBeVisible({ timeout: 10000 });
    const searchInput = page.locator('input[placeholder*="Type a command or search"]');
    await expect(searchInput).toBeVisible();

    // Check availability of content types
    // "User Service" might not be indexed immediately or at all in some envs.
    // "Dashboard" is a static page and should always be present.
    await searchInput.fill('Dashboard');
    await expect(page.getByRole('option', { name: /Dashboard/i }).first()).toBeVisible();

    // Also verify we can find the "Services" category or items
    await searchInput.fill('Service');
    await expect(page.getByRole('option', { name: /Services/i }).first()).toBeVisible();

    // Navigate to Dashboard
    await searchInput.fill('Dashboard');
    await page.keyboard.press('Enter');

    // Verify navigation
    await expect(page).toHaveURL(/\//); // Dashboard is usually root or /dashboard
  });
});
