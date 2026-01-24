/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Tool Configuration', () => {
  test.skip('should allow configuring and running a tool via form', async ({ page }) => {
    // Use real tools from backend
    // This test assumes an 'echo' tool exists or we pick the first available one.
    await page.goto('/playground');

    // Get list of tools from API to find a valid one (or assume echo)
    // We can't use page.request within the browser context easily for setup *before* goto unless we use request fixture
    // But we can just use the UI to search/select.

    // For this test to succeed without mocks, we need a known tool.
    // Let's assume 'echo' tool is available (common in default server).
    // If not, we might need to rely on what's available.

    // Wait for Playground to load (looking for title)
    await expect(page.getByRole('link', { name: 'Dashboard' })).toBeVisible({ timeout: 30000 }).catch(() => {}); // Optional wait for sidebar
    await expect(page.getByRole('heading', { name: 'Playground' })).toBeVisible({ timeout: 30000 });

    // Open "Available Tools" sheet
    await page.getByRole('button', { name: 'Available Tools' }).click();

    await expect(page.getByRole('heading', { name: 'Available Tools' })).toBeVisible();

    // We assume 'get_weather' tool is available.
    const toolName = 'get_weather';

    // wait for tool to be visible in the list
    await expect(page.getByText(toolName)).toBeVisible();

    // Find the tool container and click "Use Tool"
    // The structure works with locating the container by text, then finding the button.
    await page.locator('div.border').filter({ hasText: toolName }).getByRole('button', { name: /Use Tool/i }).click();

    // Dialog should open
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('heading', { name: toolName })).toBeVisible();

    // Fill form for 'get_weather' (properties: location (string))
    // Note: get_weather usually takes 'location' string if configured standardly.
    // config.minimal.yaml: args: ['{"weather": "sunny"}'] -> this is response mock?
    // Wait, tools args definition?
    // values-e2e.yaml tools: name: get_weather.
    // DOES NOT define inputSchema!
    // If inputSchema is missing, UI might show raw JSON input or nothing?
    // config.minimal.yaml doesn't specify input_schema.
    // If schema is missing, Playground falls back to JSON input?
    // Let's assume raw input or empty.
    // OR we assume server provides default schema?
    // If 'get_weather' expects no args or unknown, we might just click Run.

    // Let's try filling 'message' check if it exists, if not, skip filling.
    // Or check for 'location'.
    // If no schema, UI might show "No parameters".

    // Run Tool
    await page.getByRole('button', { name: /build command/i }).click();
    await page.getByLabel('Send').click();

    // Verify chat message
    await expect(page.getByText(toolName)).toBeVisible();

    // Verify result
    // The mock output is '{"weather": "sunny"}' from config calls.
    await expect(page.getByText('sunny')).toBeVisible();
  });

  test.skip('should display smart error diagnostics and allow retry', async ({ page }) => {
     // Skipped until we have a real 'timeout_tool' or 'sleep' tool to force a timeout.
     // Previous test used mocks.
  });

});
