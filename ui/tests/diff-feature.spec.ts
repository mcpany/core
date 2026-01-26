/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Tool Output Diffing', () => {
  test('should allow comparing tool outputs when they differ', async ({ page }) => {
    // Mock the tools API response
    await page.route('/api/v1/tools', async route => {
      const json = {
        tools: [
          {
            name: 'diff_test_tool',
            description: 'Test diffing',
            inputSchema: {
              type: 'object',
              properties: {
                arg: { type: 'string' }
              }
            }
          }
        ]
      };
      await route.fulfill({ json });
    });

    // Mock the tool execution
    let callCount = 0;
    await page.route('/api/v1/execute', async route => {
      callCount++;
      const result = callCount === 1 ? { value: "Version 1" } : { value: "Version 2" };

      await route.fulfill({
        json: {
          content: [
            {
              type: 'text',
              text: JSON.stringify(result)
            }
          ],
          isError: false,
          ...result
        }
      });
    });

    await page.goto('/playground');

    // 1. Run the tool first time
    await page.fill('input[placeholder="Enter command or select a tool..."]', 'diff_test_tool {"arg":"test"}');
    await page.keyboard.press('Enter');

    // Wait for first result
    await expect(page.getByText('"Version 1"')).toBeVisible();

    // 2. Run the tool second time (same args)
    // The input clears after send, so we type again.
    await page.fill('input[placeholder="Enter command or select a tool..."]', 'diff_test_tool {"arg":"test"}');
    await page.keyboard.press('Enter');

    // Wait for second result
    await expect(page.getByText('"Version 2"')).toBeVisible();

    // 3. Check for "Show Changes" button
    // It SHOULD be visible now.
    const showDiffBtn = page.getByRole('button', { name: 'Show Changes' });
    await expect(showDiffBtn).toBeVisible();

    // 4. Click the button
    await showDiffBtn.click();

    // 5. Verify Dialog opens and Diff Editor is present
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText('Output Difference')).toBeVisible();

    // Check for Monaco Diff Editor. It usually has a class 'monaco-diff-editor'.
    // Or we can check for the content text being present twice (original and modified).
    // Monaco renders text in lines.
    await expect(page.locator('.monaco-diff-editor')).toBeVisible();

    // Take verification screenshot
    await page.screenshot({ path: 'diff-feature.png' });
  });
});
