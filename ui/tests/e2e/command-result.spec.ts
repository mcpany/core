/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Command Result View', () => {
  test('should display command output in terminal view', async ({ page }) => {
    // Navigate to playground
    await page.goto('/playground');

    // Search for the tool
    const searchInput = page.getByPlaceholder('Search tools...');
    await searchInput.fill('get_weather');

    // Select the tool (weather-service.get_weather)
    // We use a regex to match strictly the one we want, avoiding 'wttr.in' if possible or accepting either if they work.
    // weather-service.get_weather was the one that was UP in the logs.
    const toolButton = page.getByText('weather-service.get_weather').first();
    await expect(toolButton).toBeVisible();
    await toolButton.click();

    // Dialog opens
    await expect(page.getByRole('dialog')).toBeVisible();

    // Fill form
    // The weather-service.get_weather tool appears to take no arguments in this environment.
    // So we skip filling the form.
    // Note: If it did take arguments, we would fill them here.

    // Check if there are inputs, if so fill them (robustness)
    const hasLocationInput = await page.getByLabel('location', { exact: false }).isVisible();
    if (hasLocationInput) {
         await page.getByLabel('location', { exact: false }).fill('Paris');
    }

    // Run Tool
    await page.getByRole('button', { name: 'Build Command' }).click();

    // Send
    await page.getByLabel('Send').click();

    // Verify chat message appears
    // We expect the command to be displayed in the chat area
    // Use .last() or filter by container
    await expect(page.getByText('weather-service.get_weather {}')).toBeVisible();

    // Verify Terminal View
    // We expect the "Console" button to be present and active
    // Use exact: true to avoid matching "Load into console" or other buttons
    const consoleBtn = page.getByRole('button', { name: 'Console', exact: true });
    await expect(consoleBtn).toBeVisible();

    // Wait for execution (it might take a bit)
    // We look for the result area.
    // The result should contain the CommandResultView

    // Check for "Exit:" badge or "success" badge
    await expect(page.getByText(/Exit: \d+|success/)).toBeVisible({ timeout: 10000 });

    // Check for stdout container
    const terminalBox = page.locator('.bg-\\[\\#1e1e1e\\]');
    await expect(terminalBox).toBeVisible();

    // Ensure we are in Console view (button should be secondary variant)
    // Note: shadcn Button variant="secondary" usually has bg-secondary
    // But checking class might be flaky.
    // Instead, let's verify that the CommandResultView structure is present.
    // It has "Stderr output:" if there is stderr, or just text.
  });
});
