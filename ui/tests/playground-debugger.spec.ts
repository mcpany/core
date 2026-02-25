/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Debugger & Presets', () => {
  // Use real backend data (weather-service.get_weather tool from config.minimal.yaml)
  test('should allow saving presets and intercepting execution', async ({ page }) => {
    await page.goto('/playground');

    // Wait for sidebar to load tools
    await expect(page.getByText('weather-service.get_weather')).toBeVisible({ timeout: 30000 });

    // Click tool card
    await page.locator('.group').filter({ hasText: 'weather-service.get_weather' }).click();

    // --- TEST PRESETS ---

    // Switch to JSON tab
    await page.getByRole('tab', { name: 'JSON' }).first().click();

    // Fill JSON arguments
    await page.getByPlaceholder('{}').fill('{"city": "Paris"}');

    // Open Presets menu
    await page.getByTitle('Manage Presets').click();

    // Click "Create New Preset" (Plus icon)
    await page.getByTitle('Create New Preset').click();

    // Enter name "Paris Weather"
    await page.getByPlaceholder('Preset Name').fill('Paris Weather');

    // Save by pressing Enter
    await page.getByPlaceholder('Preset Name').press('Enter');

    // Verify preset list updated
    // Use more specific selector to avoid matching toast
    await expect(page.locator('span.truncate').filter({ hasText: 'Paris Weather' })).toBeVisible();

    // Close popover by clicking on the preset (which also loads it, but that's fine)
    await page.locator('span.truncate').filter({ hasText: 'Paris Weather' }).click();

    // Clear input manually to ensure we can restore it
    await page.getByPlaceholder('{}').fill('{}');

    // Open presets again
    await page.getByTitle('Manage Presets').click();

    // Load preset
    await page.locator('span.truncate').filter({ hasText: 'Paris Weather' }).click();

    // Verify input restored
    await expect(page.getByPlaceholder('{}')).toHaveValue(/\"city\": \"Paris\"/);


    // --- TEST INTERCEPTOR ---

    // Enable Interceptor
    await page.getByTitle('Interceptor Mode (Breakpoint)').click();

    // Execute
    await page.getByRole('button', { name: 'Execute' }).click();

    // Verify Interceptor Dialog opened
    await expect(page.getByRole('heading', { name: 'Breakpoint Hit: weather-service.get_weather' })).toBeVisible();

    // Verify payload in dialog
    // The dialog textarea should contain the payload
    await expect(page.locator('div[role="dialog"] textarea')).toHaveValue(/\"city\": \"Paris\"/);

    // Modify payload in interception dialog
    await page.locator('div[role="dialog"] textarea').fill('{\n  "city": "Tokyo"\n}');

    // Resume Execution
    await page.getByRole('button', { name: 'Resume Execution' }).click();

    // Verify Dialog closed
    await expect(page.getByRole('heading', { name: 'Breakpoint Hit' })).not.toBeVisible();

    // Verify Result Success
    await expect(page.getByText('Success')).toBeVisible({ timeout: 10000 });
  });
});
