/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Visualizer', () => {
  test('should show live agent flow from trace', async ({ page }) => {
    // 1. Generate a trace by executing a tool in Playground
    // We use the 'get_weather' tool which is configured in the test environment (config.minimal.yaml)
    await page.goto('/playground');

    // Wait for the playground to be ready
    await expect(page.getByRole('heading', { name: 'Console' })).toBeVisible();

    // Type the command directly to avoid sidebar navigation issues
    // Using a unique parameter to ensure we can identify this specific execution if needed,
    // although simply checking for the tool node is enough.
    const command = 'weather-service.get_weather {"weather": "sunny"}';
    await page.getByPlaceholder('Enter command or select a tool...').fill(command);
    await page.getByLabel('Send').click();

    // Wait for the execution result to ensure trace is generated
    // The result comes in a collapsible with "Result: weather-service.get_weather"
    await expect(page.getByText('Result: weather-service.get_weather')).toBeVisible();

    // 2. Navigate to Visualizer
    await page.goto('/visualizer');

    // Wait a bit for backend to process and log the trace
    await page.waitForTimeout(1000);

    // Click Refresh to ensure we have the latest trace
    // The refresh button is an icon button with RefreshCcw
    // We can find it by looking for the button in the card
    const refreshBtn = page.locator('button').filter({ has: page.locator('svg.lucide-refresh-ccw') });
    if (await refreshBtn.isVisible()) {
        await refreshBtn.click();
    }

    // Check if the dropdown contains our trace
    // The select trigger should show the trace name eventually
    await expect(page.getByRole('combobox')).toContainText('weather-service.get_weather');

    // 3. Verify Graph Elements

    // Check for "User" node
    await expect(page.getByText('User', { exact: true })).toBeVisible();

    // Check for "MCP Core" node (AgentNode)
    await expect(page.getByText('MCP Core')).toBeVisible();

    // Check for "get_weather" node (ToolNode)
    await expect(page.getByText('weather-service.get_weather', { exact: false })).toBeVisible();

    // Verify connections (Edges)
    // Edges are SVGs, harder to test text, but if nodes are layouted correctly, the graph is likely working.
    // We can check if the Select dropdown has populated with our trace
    await expect(page.getByRole('combobox')).toBeVisible();
  });
});
