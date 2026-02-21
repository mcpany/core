/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Tool Configuration', () => {
  test('should allow configuring and running a tool via runner', async ({ page }) => {
    // Navigate to playground
    await page.goto('/playground');

    // Wait for real tools to load (no mocks)
    // We expect 'get_weather' from config.minimal.yaml
    await expect(page.getByText('get_weather')).toBeVisible({ timeout: 15000 });

    // Click "Use" on get_weather
    // Find the card containing 'get_weather'
    const toolCard = page.locator('.group', { hasText: 'get_weather' }).first();
    // Hover to reveal button if needed (opacity transition in CSS)
    await toolCard.hover();
    await toolCard.getByRole('button', { name: 'Use' }).click();

    // Verify "Tool Runner" tab is active
    await expect(page.getByRole('tab', { name: 'Tool Runner' })).toHaveAttribute('data-state', 'active');

    // Verify Tool Runner header shows tool name
    await expect(page.getByRole('heading', { name: 'get_weather' })).toBeVisible();

    // Switch to JSON view to be safe (in case schema is missing/empty)
    // Note: ToolRunner tabs are "Form" and "JSON" inside "Test & Execute"
    // We target the tab list inside the runner
    await page.getByRole('tab', { name: 'JSON' }).click();

    // Type some JSON
    const input = page.locator('textarea').first();
    await input.fill('{"test": "value"}');

    // Click Run (in Tool Runner)
    await page.getByRole('button', { name: 'Run' }).click();

    // Verify Result
    // Wait for the empty state to disappear
    await expect(page.getByText('Execute to see results...')).not.toBeVisible({ timeout: 10000 });

    // Verify we got a result (echo should return success)
    // We can check for "Result" label presence which is always there,
    // but better to check if content rendered.
    // RichResultViewer usually renders "content" or "text".
    // Or we can check if "Error" is NOT visible.
    await expect(page.getByText('Error:')).not.toBeVisible();
  });
});
