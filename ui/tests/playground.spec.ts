/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground', () => {
  test('should execute calculator tool', async ({ page }) => {
    // Go to playground
    await page.goto('/playground');

    // Wait for initial message
    await expect(page.getByText('Hello! I am your MCP Assistant')).toBeVisible();

    // Type command
    const input = page.locator('input[placeholder*="calculator"]');
    await input.fill('calculator {"operation": "add", "a": 10, "b": 20}');
    await input.press('Enter');

    // Wait for result
    await expect(page.getByText('Result (calculator)')).toBeVisible({ timeout: 10000 });

    // Check for the result value "Mock execution result" in the JSON output
    await expect(page.locator('pre').filter({ hasText: 'Mock execution result' })).toBeVisible();
  });

  test('should show error for invalid json', async ({ page }) => {
    await page.goto('/playground');

    const input = page.locator('input[placeholder*="calculator"]');
    await input.fill('calculator { invalid json }');
    await input.press('Enter');

    await expect(page.getByText('Invalid JSON arguments')).toBeVisible();
  });
});
