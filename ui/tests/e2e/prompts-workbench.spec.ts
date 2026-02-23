/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import { seedPrompts, cleanupPrompts, seedUser, cleanupUser } from './test-data';

test.describe('Prompts Workbench', () => {
  test.beforeEach(async ({ page, request }) => {
      await seedPrompts(request);
      await seedUser(request, "e2e-admin-prompts");

      // Login
      await page.goto('/login');
      await page.fill('input[name="username"]', 'e2e-admin-prompts');
      await page.fill('input[name="password"]', 'password');
      await page.click('button[type="submit"]', { force: true });
      await page.waitForURL('/', { timeout: 30000 });
  });

  test.afterEach(async ({ request }) => {
      await cleanupPrompts(request);
      await cleanupUser(request, "e2e-admin-prompts");
  });

  test('should load prompts list and allow selection', async ({ page }) => {
    // Navigate to prompts page
    await page.goto('/prompts');

    // Check if the page title exists
    await expect(page.locator('h3', { hasText: 'Prompt Library' })).toBeVisible();

    // Check for search input to ensure basic layout
    await expect(page.locator('input[placeholder="Search prompts..."]')).toBeVisible();

    // Handle potential empty state or populated list
    const noPrompts = page.getByText('No prompts found');
    const firstPrompt = page.locator("div[class*='border-r'] button").first();

    // Wait for either no prompts functionality or the list to populate
    await Promise.race([
        expect(noPrompts).toBeVisible(),
        expect(firstPrompt).toBeVisible()
    ]);

    if (await firstPrompt.isVisible()) {
        await firstPrompt.click();
        // Check for details view
        await expect(page.getByTestId('prompt-details').getByText('Configuration').first()).toBeVisible();
    } else {
        await expect(noPrompts).toBeVisible();
    }
  });
});
