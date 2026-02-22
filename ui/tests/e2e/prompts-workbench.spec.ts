/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedPrompts, cleanupPrompts } from './test-data';

test.describe('Prompts Workbench', () => {
  test.beforeEach(async ({ request }) => {
    await seedPrompts(request);
  });

  test.afterEach(async ({ request }) => {
    await cleanupPrompts(request);
  });

  test('should load prompts list and allow selection', async ({ page }) => {
    // Navigate to prompts page
    await page.goto('/prompts');

    // Check if the page title exists
    await expect(page.locator('h3', { hasText: 'Prompt Library' })).toBeVisible();

    // Check for search input to ensure basic layout
    await expect(page.locator('input[placeholder="Search prompts..."]')).toBeVisible();

    // With real seeded data, we expect prompts to be present.
    // Wait for the prompt we seeded: "test-prompt"
    const testPrompt = page.locator("div[class*='border-r'] button", { hasText: 'test-prompt' }).first();
    await expect(testPrompt).toBeVisible();

    await testPrompt.click();
    // Check for details view
    await expect(page.getByTestId('prompt-details').getByText('Configuration').first()).toBeVisible();
    await expect(page.getByText('A test prompt')).toBeVisible();
  });
});
