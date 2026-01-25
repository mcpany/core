/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Tool Diffing', () => {
  test('should show diff when running tool twice with same args', async ({ page }) => {
    // Mock tools
    await page.route('/api/v1/tools', async route => {
      await route.fulfill({
        json: {
          tools: [{ name: 'diff_tool', description: 'Diff tool', inputSchema: { type: 'object', properties: {} } }]
        }
      });
    });

    let executionCount = 0;
    await page.route('/api/v1/execute', async route => {
      executionCount++;
      const response = executionCount === 1
        ? { status: "ok", version: 1 }
        : { status: "ok", version: 2 };
      await route.fulfill({ json: response });
    });

    await page.goto('/playground');
    await expect(page.getByText('diff_tool')).toBeVisible();

    // First execution
    await page.getByRole('button', { name: 'Use', exact: true }).click();
    // Empty args, so just build
    await page.getByRole('button', { name: /build command/i }).click();
    await page.getByLabel('Send').click();

    // Wait for first result
    await expect(page.getByText('"version": 1')).toBeVisible();

    // Second execution (same args - empty)
    await page.getByRole('button', { name: 'Use', exact: true }).click();
    await page.getByRole('button', { name: /build command/i }).click();
    await page.getByLabel('Send').click();

    // Wait for second result
    await expect(page.getByText('"version": 2')).toBeVisible();

    // Check for "Compare Previous" button
    // It might be in the second result card.
    // We can scope to the last result card if needed, but text uniqueness helps.
    const compareBtn = page.getByRole('button', { name: /Compare Previous/i }).last();
    await expect(compareBtn).toBeVisible();

    // Click compare
    await compareBtn.click();

    // Check for diff content
    // We expect "version": 1 to be removed (red) and "version": 2 to be added (green)
    // The implementation uses diff.diffJson which handles object comparison.
    // The rendered HTML uses Tailwind classes.

    // Note: diffJson results depend on formatting.
    // Ideally we look for red text containing "1" and green text containing "2".
    await expect(page.locator('.text-red-400')).toContainText('1');
    await expect(page.locator('.text-green-400')).toContainText('2');
  });
});
