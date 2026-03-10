/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Tool Output Diffing', () => {
  // We reset the difftest file in case it was left over from a previous test run
  test.beforeEach(async ({ request }) => {
     // A trick to reset state: we could call a tool to delete the file, or just let it toggle.
     // To ensure deterministic behavior, we run it once, check the output. If it's Version 2, we run it again so the next is Version 1.
     // But it's easier to just run it twice and check that the "Show Changes" button appears.
  });

  test('should allow comparing tool outputs when they differ', async ({ page }) => {
    // Navigate to playground
    await page.goto('/playground');

    // 1. Run the tool first time
    await page.fill('input[placeholder="Enter command or select a tool..."]', 'diff-test-service.diff_test_tool {"arg":"test"}');
    await page.keyboard.press('Enter');

    // Wait for the first result to appear (we don't know if it will be Version 1 or Version 2 based on system state, but it will be one of them)
    await expect(page.locator('.whitespace-pre-wrap', { hasText: 'Version' }).first()).toBeVisible({ timeout: 10000 });

    // 2. Run the tool second time (same args)
    // The input clears after send, so we type again.
    await page.fill('input[placeholder="Enter command or select a tool..."]', 'diff-test-service.diff_test_tool {"arg":"test"}');
    await page.keyboard.press('Enter');

    // Wait for second result
    // We expect there to be a total of 2 execution results
    await expect(page.getByText('Result: diff-test-service.diff_test_tool')).toHaveCount(2, { timeout: 10000 });

    // 3. Check for "Show Changes" button
    // It SHOULD be visible now on the second result because the output toggles.
    const showDiffBtn = page.getByRole('button', { name: 'Show Changes' });
    await expect(showDiffBtn).toBeVisible();

    // 4. Click the button
    await showDiffBtn.click();

    // 5. Verify Dialog opens and Diff Editor is present
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText('Output Difference')).toBeVisible();

    // Check for Monaco Diff Editor.
    await expect(page.locator('.monaco-diff-editor')).toBeVisible();
  });
});
