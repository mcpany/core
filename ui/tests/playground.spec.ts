/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Tool Configuration', () => {
  // Use real backend data (weather-service.get_weather tool from config.minimal.yaml)
  test('should allow configuring and running a tool via inline wizard', async ({ page }) => {
    await page.goto('/playground');

    // Wait for sidebar to load tools (real data)
    // We get the first element to avoid strict mode violations if it appears in multiple places
    await expect(page.getByText('weather-service.get_weather').first()).toBeVisible({ timeout: 30000 });

    // Ensure we are in "Console" mode
    await expect(page.getByRole('tab', { name: 'Console' })).toHaveAttribute('data-state', 'active');

    // Click the chat input box and type part of the tool name
    await page.getByPlaceholder('Enter command or select a tool...').fill('get_weather');

    // Wait for the autocomplete dropdown
    await page.waitForSelector('.absolute.bottom-full.left-0');

    // Click the autocomplete suggestion for weather-service.get_weather
    // Ensure we click the element inside the dropdown that sets activeInlineTool
    await page.locator('.absolute.bottom-full.left-0').getByText('weather-service.get_weather', { exact: true }).first().click();

    // Wait for the tool execution UI to be ready
    await page.waitForTimeout(2000);

    // Verify the inline form appears
    await expect(page.locator('h3:has-text("Configure")')).toBeVisible();

    // Switch to the JSON tab
    await page.getByRole('tab', { name: 'JSON' }).first().click();

    // Fill JSON arguments
    await page.getByPlaceholder('{}').first().fill('{"city": "San Francisco"}');

    // Click "Execute Tool" or "Execute"
    await page.getByRole('button', { name: /Execute/ }).first().click();

    // Verify "Result: weather-service.get_weather" appears in the chat stream
    await expect(page.locator('.text-green-700, .text-green-400').filter({ hasText: /Result:/ }).first()).toBeVisible({ timeout: 10000 });
  });
});
