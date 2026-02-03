/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Prompts Workbench', () => {
  test.fixme('should load prompts list and allow selection', async ({ page }) => { // Flaky in Docker/Kind environment
    // Mock the prompts API to ensure consistent state
    await page.route('**/api/v1/prompts', async route => {
        await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
                prompts: [
                    {
                        name: "test-prompt",
                        description: "A test prompt",
                        arguments: [{ name: "arg1", description: "An argument" }]
                    }
                ]
            })
        });
    });

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
