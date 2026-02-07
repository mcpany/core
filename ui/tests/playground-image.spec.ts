/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Image Rendering', () => {

  const SERVICE_NAME = 'image-test-service';

  test.beforeAll(async ({ request }) => {
    // Clean up if exists
    await request.delete(`/api/v1/services/${SERVICE_NAME}`).catch(() => {});

    // Register a command line service
    const response = await request.post('/api/v1/services', {
      data: {
        id: SERVICE_NAME,
        name: SERVICE_NAME,
        command_line_service: {
            tools: [
                {
                    name: 'get_image',
                    description: 'Returns a red dot image',
                    call_id: 'get_image_call',
                    title: 'Get Red Dot Image'
                }
            ],
            calls: {
                'get_image_call': {
                    // We use 'echo' to output the JSON structure of a CallToolResult content array.
                    // The Command tool wraps stdout in `stdout` field of the result.
                    // SmartResultRenderer unwraps `stdout` if it is valid JSON.
                    args: [
                        '[{"type": "image", "data": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=", "mimeType": "image/png"}]'
                    ]
                }
            },
            command: "echo"
        }
      }
    });
    if (!response.ok()) {
        console.error('Failed to register service:', await response.text());
    }
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    // Cleanup
    await request.delete(`/api/v1/services/${SERVICE_NAME}`).catch(() => {});
  });

  test('should render image result from tool', async ({ page }) => {
    // Navigate to playground with specific tool selected to speed up test
    // But let's follow the user flow
    await page.goto('/playground');

    // Wait for tool list to load (real data)
    // Sometimes the list takes a moment to refresh or might be cached.
    // We might need to refresh if not found?
    // The sidebar loads via `apiClient.listTools()`.
    await expect(page.getByText('get_image')).toBeVisible({ timeout: 10000 });

    // Select tool
    await page.getByText('get_image').click();

    // Wait for dialog
    await expect(page.getByRole('dialog')).toBeVisible();

    // Click "Run" (or Build Command -> Run)
    // Try to find the primary action button in the dialog footer
    const buildBtn = page.getByRole('button', { name: /build command/i });
    const runBtn = page.getByRole('button', { name: /run/i });
    const insertBtn = page.getByRole('button', { name: /insert/i });

    if (await buildBtn.isVisible()) {
        await buildBtn.click();
    } else if (await insertBtn.isVisible()) {
        await insertBtn.click();
    } else if (await runBtn.isVisible()) {
        await runBtn.click();
    } else {
        // Fallback: click the last button in dialog
        const dialog = page.getByRole('dialog');
        await dialog.getByRole('button').last().click();
    }

    // Dialog closes, input populated.
    // Click "Send"
    const sendBtn = page.getByLabel('Send');
    await expect(sendBtn).toBeEnabled();
    await sendBtn.click();

    // Wait for result
    // The image should appear.
    // Selector: img[src^="data:image/png"]
    const img = page.locator('img[src^="data:image/png"]');
    await expect(img).toBeVisible({ timeout: 15000 });
  });
});
