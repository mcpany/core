/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Image Rendering', () => {
  const serviceName = 'image-test-service-' + Date.now();
  const toolName = 'generate_image';

  // A 1x1 red pixel PNG base64
  const base64Image = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=';
  const jsonOutput = JSON.stringify({
    content: [
      {
        type: 'image',
        data: base64Image,
        mimeType: 'image/png'
      }
    ]
  });

  test.beforeAll(async ({ playwright }) => {
    const request = await playwright.request.newContext();
    // Register a command line service
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        call_policies: [
            {
                default_action: 0 // ALLOW
            }
        ],
        command_line_service: {
          command: 'echo',
          local: true,
          tools: [
            {
              name: toolName,
              call_id: 'call1',
              description: 'Generate a red pixel'
            }
          ],
          calls: {
            'call1': {
              args: [jsonOutput]
            }
          }
        }
      },
      headers: {
          'Content-Type': 'application/json',
          'X-API-Key': process.env.MCPANY_API_KEY || ''
      }
    });

    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ playwright }) => {
      const request = await playwright.request.newContext();
      // Cleanup
      await request.delete(`/api/v1/services/${serviceName}`, {
          headers: {
              'X-API-Key': process.env.MCPANY_API_KEY || ''
          }
      });
  });

  test('should execute tool and display image result', async ({ page }) => {
    await page.goto('/playground');

    // Wait for services to load and finding our service might take a moment if list is long
    // But playground loads all tools.

    // Open sidebar if it's not open (desktop default is open)
    // We can just search for the tool in the sidebar
    await page.getByPlaceholder('Search tools...').fill(toolName);

    // Click on the tool in the sidebar
    await page.getByText(toolName).click();

    // Dialog configuration should appear
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('heading', { name: toolName })).toBeVisible();

    // Click "Run Tool" (since there are no args, it might just be "Run" or "Build Command" then "Send")
    // The UI usually shows "Build Command" then puts it in the input.
    await page.getByRole('button', { name: /build command/i }).click();

    // Now click Send in the console
    await page.getByLabel('Send').click();

    // Wait for the image to appear
    // The src should start with data:image/png;base64,
    const imgSelector = `img[src="data:image/png;base64,${base64Image}"]`;
    await expect(page.locator(imgSelector)).toBeVisible({ timeout: 10000 });

    // Verify specific text from SmartResultRenderer (mime type label)
    await expect(page.getByText('image/png')).toBeVisible();
  });
});
