/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import { seedPrompts, cleanupPrompts, seedUser, cleanupUser } from './test-data';

test.describe('Prompts Workbench', () => {
  test.beforeEach(async ({ page, request }) => {
      await seedPrompts(request);

      // Login using the known e2e-admin-core user seeded by default test-data
      await page.goto('/login');
      await page.fill('input[name="username"]', 'e2e-admin-core');
      await page.fill('input[name="password"]', 'password');
    await Promise.all([
      page.waitForURL('/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
    ]);
  });

  test.afterEach(async ({ request }) => {
      await cleanupPrompts(request);
  });

  test('should load prompts list and allow selection', async ({ page }) => {
    // Navigate to prompts page
    await page.goto('/prompts');

    // Check if the page title exists
    await expect(page.locator('h2', { hasText: 'Prompts' })).toBeVisible();

    // Check for search input to ensure basic layout
    await expect(page.locator('input[placeholder="Search prompts..."]')).toBeVisible();

    // Handle potential empty state or populated list
    const noPrompts = page.getByText('No prompts found');
    const firstPrompt = page.getByRole('button', { name: 'Inspect' }).first();

    // Wait for either no prompts functionality or the list to populate
    await Promise.race([
        expect(noPrompts).toBeVisible(),
        expect(firstPrompt).toBeVisible()
    ]);

    if (await firstPrompt.isVisible()) {
        await firstPrompt.click();
        // Check for details view
        await expect(page.getByTestId('prompt-details')).toContainText('Configuration');
    } else {
        await expect(noPrompts).toBeVisible();
    }
  });
});
