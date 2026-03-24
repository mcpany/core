/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Basic Verification', () => {
  test('should execute calculator tool and verify output', async ({ page }) => {
    // Mock the tools API
    await page.route('**/api/v1/tools', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          tools: [
            {
              name: 'calculator',
              description: 'A simple calculator',
              inputSchema: {
                type: 'object',
                properties: {
                  operation: { type: 'string', enum: ['add', 'subtract'] },
                  a: { type: 'number' },
                  b: { type: 'number' }
                },
                required: ['operation', 'a', 'b']
              }
            }
          ]
        })
      });
    });

    // Mock the execute API
    await page.route('**/api/v1/execute', async (route) => {
        const body = JSON.parse(route.request().postData() || '{}');
        await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
                content: [{ type: "text", text: `Result: ${body.arguments.a + body.arguments.b}` }]
            })
        });
    });

    // Navigate to playground
    await page.goto('/playground');

    // Waiting for chat input
    const chatInput = page.getByPlaceholder('Enter command or select a tool...');
    await expect(chatInput).toBeVisible({ timeout: 10000 });

    // Type a command
    const msg = 'calculator {"operation": "add", "a": 5, "b": 3}';
    await chatInput.fill(msg);

    // Click Send
    const sendBtn = page.getByRole('button', { name: 'Send' });
    await expect(sendBtn).toBeVisible();
    await sendBtn.click();

    // Assert: Check if message appears
    await expect(page.getByText(msg)).toBeVisible({ timeout: 10000 });

    // Checking layout (Library visible)
    await expect(page.getByText('Library')).toBeVisible();

    // Verify result (Execution happened)
    await expect(page.getByText('Result: 8')).toBeVisible();
  });
});
