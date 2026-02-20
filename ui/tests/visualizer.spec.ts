/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Visualizer E2E', () => {
  test('should display live trace after tool execution', async ({ page }) => {
    // 1. Go to Playground
    await page.goto('/playground');

    // 2. Execute a tool to generate a trace
    // Wait for the tool list to load
    await expect(page.getByText('get_weather')).toBeVisible({ timeout: 10000 });

    // Open tool dialog
    // Assuming the sidebar is open or we can click it
    // The "Use" button is usually on the tool card/item in the sidebar
    // Let's find the tool item and click it
    await page.getByText('get_weather').click();

    // In PlaygroundClientPro, clicking a tool in sidebar opens the configure dialog.
    // Wait for dialog
    await expect(page.getByRole('dialog')).toBeVisible();

    // Click "Build Command" (submits the form)
    await page.getByRole('button', { name: /build command/i }).click();

    // The command is now in the input. Click Send.
    await page.getByLabel('Send').click();

    // Wait for result
    // The tool returns {"weather": "sunny"} based on config.minimal.yaml
    await expect(page.getByText('weather', { exact: false })).toBeVisible();
    await expect(page.getByText('sunny', { exact: false })).toBeVisible();

    // 3. Navigate to Visualizer
    await page.goto('/visualizer');

    // 4. Verify Graph
    // We expect nodes: "User", "MCP Core", "get_weather"
    // My custom nodes render the label.
    // ReactFlow nodes are divs with text.

    // Wait for graph to load (polling every 3s)
    // It might take up to 3s for the trace to appear.
    await expect(page.getByText('User')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('MCP Core')).toBeVisible();

    // "get_weather" node should be there.
    // Note: The sanitization logic might affect ID but label should be "get_weather"
    // The Node component renders `data.label`.
    await expect(page.getByText('get_weather')).toBeVisible();

    // Check edges?
    // ReactFlow edges are SVG paths, harder to assert text on.
    // But checking nodes exists proves the graph was built from the trace.
  });
});
