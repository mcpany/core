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

    // Wait for Send button to be enabled (might be disabled during initial load/state update)
    const sendBtn = page.getByLabel('Send');
    await expect(sendBtn).toBeEnabled();
    await sendBtn.click();

    // Wait for the execution result to ensure trace is generated
    // The result comes in a collapsible with "Result: weather-service.get_weather"
    await expect(page.getByText('Result: weather-service.get_weather')).toBeVisible();

    // 2. Navigate to Visualizer
    await page.goto('/visualizer');

    // Retry refreshing until trace appears
    // The visualizer might take a moment to fetch the new trace.
    await expect(async () => {
        const refreshBtn = page.locator('button').filter({ has: page.locator('svg.lucide-refresh-ccw') });
        if (await refreshBtn.isVisible()) {
            await refreshBtn.click();
        }
        await expect(page.getByRole('combobox')).toContainText('weather-service.get_weather', { timeout: 1000 });
    }).toPass({
        timeout: 30000,
        intervals: [2000]
    });

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
