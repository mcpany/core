/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Collections (Real Data)', () => {
  // This test assumes the backend is running with config.minimal.yaml which seeds 'weather-service'.
  // We do NOT mock any API calls to ensure "Real Data" requirement is met.

  test('should allow saving requests to a collection and replaying them', async ({ page }) => {
    // 1. Navigate to Playground
    await page.goto('/playground');

    // 2. Wait for 'get_weather' tool to appear
    // This confirms backend connection and tool listing
    await expect(page.getByText('get_weather')).toBeVisible({ timeout: 15000 });

    // 3. Select the tool
    // If there are many tools, search helps
    await page.getByPlaceholder('Search tools...').fill('get_weather');
    await page.getByText('get_weather').click();

    // 4. Build Command
    // Since inputSchema might be empty in minimal config, we just proceed.
    // If it requires args, we might need to handle that, but assuming 'echo' works with empty.
    await page.getByRole('button', { name: /build command/i }).click();

    // 5. Execute Tool
    const sendBtn = page.getByLabel('Send');
    await expect(sendBtn).toBeEnabled();
    await sendBtn.click();

    // 6. Verify Execution Result
    // We expect a tool call and a result message
    await expect(page.getByText('Tool Execution')).toBeVisible();
    await expect(page.getByText('get_weather')).toBeVisible();
    // Result should appear (even if empty or error, it returns a Result message)
    await expect(page.getByText('Result: get_weather')).toBeVisible();

    // 7. Save to Collection
    // The save button is on the tool call card header.
    // We need to target the specific tool call we just made.
    const toolCallHeader = page.locator('.flex.items-center', { hasText: 'Tool Execution' }).last();
    // Hover to reveal button? The tooltip triggers on hover, but button is always rendered conditionally?
    // In chat-message.tsx: {onSave && ... Button}
    // It should be visible if onSave prop is passed.
    await expect(page.getByLabel('Save to collection')).toBeVisible();
    await page.getByLabel('Save to collection').click();

    // 8. Configure Save Dialog
    await expect(page.getByRole('dialog', { name: 'Save Request' })).toBeVisible();
    await page.getByLabel('Name').fill('My Weather Test');
    // Note: The logic auto-creates "My Collection" if none exists.
    await page.getByRole('button', { name: 'Save' }).click();

    // 9. Verify Collections Sidebar
    // Switch tab
    await page.getByRole('button', { name: 'Collections' }).click();

    // Check Collection exists
    await expect(page.getByText('My Collection')).toBeVisible();

    // Expand/Check Request exists
    // It might be auto-expanded
    await expect(page.getByText('My Weather Test')).toBeVisible();

    // 10. Run from Collection
    // First clear the chat to be sure
    await page.getByRole('button', { name: 'Clear' }).click();
    await expect(page.queryByText('Result: get_weather')).toBeNull();

    // Click Run on the saved request item
    // The item has the name "My Weather Test". We need to find the play button inside its row.
    const requestRow = page.locator('div', { hasText: 'My Weather Test' }).first();
    // We need to hover to see the button?
    await requestRow.hover();
    await requestRow.locator('button[title="Run"]').click();

    // 11. Verify Re-execution
    await expect(page.getByText('Result: get_weather')).toBeVisible();
  });
});
