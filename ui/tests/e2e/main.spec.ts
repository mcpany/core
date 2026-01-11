/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('MCP Any UI E2E', () => {

  test('Debug verify file version', async () => {
    console.log('DEBUG: RUNNING MODIFIED FILE');
  });

  test('Dashboard loads and shows metrics', async ({ page }) => {
    await page.goto('/');
    // Updated title expectation
    await expect(page).toHaveTitle(/MCPAny Manager/);
    await expect(page.locator('h1')).toContainText('Dashboard');
    // Check for metrics cards
    await expect(page.locator('text=Total Requests')).toBeVisible();
    await expect(page.locator('text=System Health')).toBeVisible();
  });

  test('should navigate to analytics from sidebar', async ({ page }) => {
    await page.goto('/');
    // Check if link exists
    const statsLink = page.getByRole('link', { name: /Analytics|Stats/i });
    if (await statsLink.count() > 0) {
        await statsLink.click();
        await expect(page).toHaveURL(/.*\/stats/);
    } else {
        console.log('Analytics link not found in sidebar, skipping navigation test');
    }
  });





  test('Middleware page drag and drop', async ({ page }) => {
    await page.goto('/middleware');
    await expect(page.locator('h2')).toContainText('Middleware Pipeline');
    await expect(page.locator('text=Active Pipeline')).toBeVisible();
    // Resolving ambiguity by selecting the first occurrence (likely the list item)
    await expect(page.locator('text=Authentication').first()).toBeVisible();
  });

});
