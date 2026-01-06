/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Global Search', () => {


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
