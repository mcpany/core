/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Tool Output Diffing', () => {
<<<<<<< HEAD
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
=======
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
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
    const showDiffBtn = page.getByRole('button', { name: 'Show Changes' });
    await expect(showDiffBtn).toBeVisible();

    // 4. Click the button
    await showDiffBtn.click();

    // 5. Verify Dialog opens and Diff Editor is present
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText('Output Difference')).toBeVisible();
<<<<<<< HEAD
    await expect(page.locator('.monaco-diff-editor')).toBeVisible();
=======

    // Check for Monaco Diff Editor. It usually has a class 'monaco-diff-editor'.
    // Or we can check for the content text being present twice (original and modified).
    // Monaco renders text in lines.
    await expect(page.locator('.monaco-diff-editor')).toBeVisible();



>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
  });
});
