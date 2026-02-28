/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Diff Verification', () => {
  test('should parse and display JSON nicely in diff viewer', async ({ page }) => {
    // Navigate to playground
    await page.goto('/playground');

    // Wait for the playground to load
    await expect(page.getByPlaceholder('Enter command or select a tool...')).toBeVisible();

    // Inject state manually to simulate two consecutive tool calls with same args but different stringified JSON results
    await page.evaluate(() => {
        const messages = [
            {
                "id": "1",
                "type": "tool-call",
                "toolName": "echo",
                "toolArgs": {"value": "Version 1"},
                "timestamp": new Date().toISOString()
            },
            {
                "id": "2",
                "type": "tool-result",
                "toolName": "echo",
                "toolResult": {
                    "content": [{
                        "type": "text",
                        "text": "{\"nested\": {\"value\":\"Version 1\"}}"
                    }],
                    "isError": false
                },
                "timestamp": new Date().toISOString()
            },
            {
                "id": "3",
                "type": "tool-call",
                "toolName": "echo",
                "toolArgs": {"value": "Version 2"},
                "timestamp": new Date().toISOString()
            },
            {
                "id": "4",
                "type": "tool-result",
                "toolName": "echo",
                "previousResult": {
                    "content": [{
                        "type": "text",
                        "text": "{\"nested\": {\"value\":\"Version 1\"}}"
                    }],
                    "isError": false
                },
                "toolResult": {
                    "content": [{
                        "type": "text",
                        "text": "{\"nested\": {\"value\":\"Version 2\"}}"
                    }],
                    "isError": false
                },
                "timestamp": new Date().toISOString()
            }
        ];
        window.localStorage.setItem('playground-messages', JSON.stringify(messages));
    });

    // Reload to apply injected state
    await page.goto('/playground');

    // Click "Show Changes" button
    const showChangesBtn = page.getByRole('button', { name: 'Show Changes' }).first();
    await expect(showChangesBtn).toBeVisible();
    await showChangesBtn.click();

    // The diff editor should appear
    await expect(page.getByText('Output Difference')).toBeVisible();

    // Since we applied deepParseJson, the stringified text is now parsed as a JSON object
    // We expect "nested" to appear without escaped quotes
    // Note: DiffEditor loads Monaco, which renders lines asynchronously. We look for text inside Monaco's view-lines.
    const monacoEditor = page.locator('.view-lines').first();
    await expect(monacoEditor).toBeVisible({ timeout: 10000 });

    // Check that the unescaped keys and values exist in the diff view
    // Monaco diff editor renders original and modified views
    const monacoEditorAll = page.locator('.view-lines');
    await expect(monacoEditorAll.last()).toContainText('"nested"');
    await expect(monacoEditorAll.last()).toContainText('"value"');
    await expect(monacoEditorAll.last()).toContainText('Version 2');
  });
});
