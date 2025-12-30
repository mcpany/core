/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

test.describe('Global Search', () => {
  test('should open command palette with Cmd+K and navigate', async ({ page }) => {
    // Go to the dashboard
    await page.goto('/');

    // Press Cmd+K (Meta+K on Mac, Control+K on Windows/Linux)
    // Try both meta and control to be safe in CI environment
    await page.waitForTimeout(1000);
    if (process.platform === 'darwin') {
        await page.keyboard.press('Meta+K');
    } else {
        await page.keyboard.press('Control+K');
    }

    // Wait for the command palette to open
    const commandDialog = page.locator('[role="dialog"]');
    await expect(commandDialog).toBeVisible();

    // Type "Playground" into the input
    const input = page.locator('[cmdk-input]');
    await input.fill('Playground');

    // Wait for the item to be selected/visible
    const playgroundItem = page.locator('[cmdk-item]').filter({ hasText: 'Playground' });
    await expect(playgroundItem).toBeVisible();

    // Click it (or press Enter)
    await playgroundItem.click();

    // Verify we navigated to /playground
    await expect(page).toHaveURL(/\/playground/);

    // Screenshot for audit
    await page.goto('/');
    await page.waitForTimeout(500); // Wait for hydration
    await page.keyboard.press('Control+K');
    await expect(commandDialog).toBeVisible();
    await input.fill(''); // Clear input to show all options

    const screenshotDir = path.join(process.cwd(), '.audit/ui', new Date().toISOString().split('T')[0]);
    if (!fs.existsSync(screenshotDir)) {
      fs.mkdirSync(screenshotDir, { recursive: true });
    }
    await page.screenshot({ path: path.join(screenshotDir, 'global_search.png') });
  });

  test('should open command palette by clicking the search button', async ({ page }) => {
     await page.goto('/');

     // Find the button with text "Search" or similar
     const searchButton = page.locator('button', { hasText: 'Search' }).first();
     await searchButton.click();

     const commandDialog = page.locator('[role="dialog"]');
     await expect(commandDialog).toBeVisible();
  });
});
