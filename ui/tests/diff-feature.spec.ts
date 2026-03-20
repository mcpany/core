/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Tool Output Diffing', () => {
  test('should allow comparing tool outputs when they differ using seeded state', async ({ page, request: apiReq }) => {
    const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
    const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };

    // Seed the backend with a stateful tool for real data verification
    await apiReq.post('/api/v1/services', {
        headers: HEADERS,
        data: {
            id: "diff_test_svc",
            name: "Diff Test Service",
            cmd_service: {
                tools: [{ name: "diff_test_tool", description: "Diff test", call_id: "diff_call" }],
                calls: {
                    diff_call: {
                        command: "bash",
                        args: ["-c", "if [ ! -f /tmp/diff_state.txt ]; then echo '[{\"id\":1, \"name\":\"Alice\"}]' > /tmp/diff_state.txt; else echo '[{\"id\":1, \"name\":\"Alice\"}, {\"id\":2, \"name\":\"Bob\"}]' > /tmp/diff_state.txt; fi; cat /tmp/diff_state.txt"]
                    }
                }
            }
        }
    });

    // Cleanup the state file so the test is repeatable
    await apiReq.post('/api/v1/services', {
        headers: HEADERS,
        data: {
            id: "cleanup_svc",
            name: "Cleanup Service",
            cmd_service: {
                tools: [{ name: "cleanup_tool", description: "Cleanup", call_id: "cleanup_call" }],
                calls: {
                    cleanup_call: {
                        command: "rm",
                        args: ["-f", "/tmp/diff_state.txt"]
                    }
                }
            }
        }
    });
    await apiReq.post('/api/v1/execute', {
        headers: HEADERS,
        data: { name: "cleanup_tool", arguments: {} }
    });

    // Wait a moment for tools to be registered and available in the UI
    await page.waitForTimeout(1000);

    await page.goto('/playground');

    // 1. Run the tool first time
    await page.fill('input[placeholder="Enter command or select a tool..."]', 'diff_test_tool {}');
    await page.keyboard.press('Enter');

    // Wait for first result
    await expect(page.getByText('"Alice"')).toBeVisible();

    // 2. Run the tool second time (same args)
    // The input clears after send, so we type again.
    await page.fill('input[placeholder="Enter command or select a tool..."]', 'diff_test_tool {}');
    await page.keyboard.press('Enter');

    // Wait for second result
    await expect(page.getByText('"Bob"')).toBeVisible();

    // 3. Check for "Show Changes" button
    // It SHOULD be visible now.
    const showDiffBtn = page.getByRole('button', { name: 'Show Changes' });
    await expect(showDiffBtn).toBeVisible();

    // 4. Click the button
    await showDiffBtn.click();

    // 5. Verify Dialog opens and Smart Table Diff is present
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText('Output Difference')).toBeVisible();

    // The SmartDiffRenderer renders a table for arrays of objects
    await expect(page.getByRole('table')).toBeVisible();

    // Check that Bob is highlighted as added
    const bobRow = page.getByRole('row', { name: '+ 2 "Bob"' });
    await expect(bobRow).toBeVisible();

    // Check that Alice is highlighted as unchanged
    const aliceRow = page.getByRole('row', { name: '1 "Alice"' });
    await expect(aliceRow).toBeVisible();
  });
});
