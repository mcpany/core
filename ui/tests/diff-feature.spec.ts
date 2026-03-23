/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Tool Output Diffing', () => {
  const serviceName = 'diff-test-service';

  test.beforeAll(async ({ request }) => {
    // Clean up previous test runs
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});

    // Ensure state file is removed
    const execResponse = require('child_process').execSync('rm -f /tmp/playwright_diff');

    // Seed the database via API with a stateful tool
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        command_line_service: {
          command: 'bash',
          tools: [
            {
              name: 'diff_test_tool',
              description: 'Test diffing',
              call_id: 'diff_test_tool',
              input_schema: {
                type: 'object',
                properties: {
                  arg: { type: 'string' }
                }
              }
            }
          ],
          calls: {
            'diff_test_tool': {
              args: [
                '-c',
                'if [ ! -f /tmp/playwright_diff ]; then echo "{\\"value\\":\\"Version 1\\"}"; touch /tmp/playwright_diff; else echo "{\\"value\\":\\"Version 2\\"}"; rm -f /tmp/playwright_diff; fi'
              ]
            }
          }
        }
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});
    require('child_process').execSync('rm -f /tmp/playwright_diff');
  });

  test('should allow comparing tool outputs when they differ', async ({ page }) => {
    await page.goto('/playground');

    // 1. Run the tool first time
    await page.fill('input[placeholder="Enter command or select a tool..."]', 'diff-test-service.diff_test_tool {"arg":"test"}');
    await page.keyboard.press('Enter');

    // Wait for first result
    await expect(page.getByText('"Version 1"').first()).toBeVisible({ timeout: 10000 });

    // 2. Run the tool second time (same args)
    await page.fill('input[placeholder="Enter command or select a tool..."]', 'diff-test-service.diff_test_tool {"arg":"test"}');
    await page.keyboard.press('Enter');

    // Wait for second result
    await expect(page.getByText('"Version 2"').first()).toBeVisible({ timeout: 10000 });

    // 3. Check for "Show Changes" button
    const showDiffBtn = page.getByRole('button', { name: 'Show Changes' });
    await expect(showDiffBtn).toBeVisible();

    // 4. Click the button
    await showDiffBtn.click();

    // 5. Verify Dialog opens and Diff Editor is present
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText('Output Difference')).toBeVisible();
    await expect(page.locator('.monaco-diff-editor')).toBeVisible();
  });
});
