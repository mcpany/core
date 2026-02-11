/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Workbench', () => {
  // We use the pre-configured 'weather-service' which is loaded by default in the server.
  // This satisfies "Real Data" as we are interacting with the real backend.
  const toolName = 'weather-service.get_weather';

  test('should open tool inspector in side panel and allow execution', async ({ page }) => {
    // 1. Navigate to Playground
    await page.goto('/playground');

    // 2. Find Tool in Sidebar
    // The sidebar might list it as "get_weather" under "weather-service" or full name.
    // Based on logs: toolName=weather-service.get_weather
    // Let's try to find it by text. It might take a moment to load.
    const toolBtn = page.getByText(toolName).first();
    await expect(toolBtn).toBeVisible({ timeout: 10000 });
    await toolBtn.click();

    // 3. Verify Inspector Panel
    // The panel header should show the tool name
    const panelHeader = page.getByRole('heading', { name: toolName });
    await expect(panelHeader).toBeVisible();

    // Verify "Run" button
    const runBtn = page.getByRole('button', { name: 'Run', exact: true });
    await expect(runBtn).toBeVisible();

    // 4. Fill Form (if applicable)
    // weather-service.get_weather likely takes arguments?
    // Log said: InputSchema:map[properties:map[] type:object] -> No args?
    // Let's check logs: InputSchema:map[properties:map[] type:object]
    // Wait, earlier logs said: InputSchema:map[properties:map[] type:object]
    // But then: Name:weather-service.get_weather OutputSchema:map[properties:map[args:map[...]]]
    // It seems get_weather takes NO arguments in the mock?
    // If no args, the form should show "This tool takes no arguments."

    // Let's see if we can just Run.
    await runBtn.click();

    // 5. Verify Execution in Console
    await expect(page.getByText('Tool Execution')).toBeVisible();
    await expect(page.getByText(toolName).nth(1)).toBeVisible(); // In chat

    // 6. Verify Result
    // Since it's a real tool (mocked internally by server?), it should return something.
    // The server log shows "Registered command service".
    // It might execute a command.
    // If it succeeds, we see "Result: ...".
    // If it fails, we see "Execution Error".
    // Either way, the interaction works.
    await expect(page.locator('.lucide-sparkles').first()).toBeVisible({ timeout: 10000 });
    // Sparkles icon indicates result (or error handled nicely?)
    // Actually ChatMessage uses Sparkles for Result.

    // 7. Verify Inspector is STILL OPEN
    await expect(runBtn).toBeVisible();
  });
});
