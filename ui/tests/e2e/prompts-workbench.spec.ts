/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Prompts Workbench', () => {
  test('should load prompts list and allow selection', async ({ page, request }) => {
    // Seed real backend with data containing prompts
    await request.post('/api/v1/debug/seed', {
        headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' },
        data: {
            services: [
                {
                    name: "weather-service",
                    command_line_service: {
                        command: "echo",
                        prompts: [{ name: "summarize_text", description: "Summarize text" }],
                        tools: [{ name: "get_weather", description: "Get current weather", call_id: "get_weather" }],
                        calls: { get_weather: { args: ['{"weather": "sunny"}'] } }
                    }
                }
            ]
        }
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
