/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Image Rendering', () => {
  const serviceName = `image-test-${Date.now()}`;

  test.beforeAll(async ({ request }) => {
    // Register a command line service that outputs an image
    // We use direct API call to bypass potential limitations in client.ts helper
    // and to ensure we send the exact structure required by the backend.
    const response = await request.post('/api/v1/services', {
      data: {
        id: serviceName,
        name: serviceName,
        command_line_service: {
          command: 'sh',
          local: true,
          tools: [
            {
              name: 'get_image',
              call_id: 'call_1',
              description: 'Returns an image'
            }
          ],
          calls: {
            'call_1': {
              id: 'call_1',
              // Use printf to avoid some echo quoting issues across platforms, though sh should be fine.
              // We return a JSON structure that SmartResultRenderer will parse from stdout.
              args: [
                '-c',
                `echo '{"content": [{"type": "image", "data": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=", "mimeType": "image/png"}]}'`
              ]
            }
          }
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
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should display image tool result', async ({ page }) => {
    await page.goto('/playground');

    // Wait for the page to load
    await expect(page.getByRole('heading', { name: 'Console' })).toBeVisible();

    // Type the command directly into the input
    // Format: tool_name {args}
    // Since get_image takes no args (schema allows anything if not defined), we can just type the name
    // or name {}.
    // Verify tool exists in sidebar
    await page.getByPlaceholder('Search tools...').fill('get_image');
    await expect(page.getByText('get_image')).toBeVisible();

    const input = page.locator('input[placeholder="Enter command or select a tool..."]');
    // Tools are namespaced by service name
    const fullToolName = `${serviceName}.get_image`;
    await input.fill(`${fullToolName} {}`);
    await input.press('Enter');

    // Wait for execution result
    // Look for the image. The src contains the base64 data.
    const imageSelector = 'img[src*="iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII="]';

    // We expect the image to appear. It might take a moment.
    await expect(page.locator(imageSelector)).toBeVisible({ timeout: 10000 });

    // Also verify we see the "Image 1" label from RichContentRenderer
    await expect(page.getByText('Image 1')).toBeVisible();
    await expect(page.getByText('image/png')).toBeVisible();

    // Take verification screenshot
    await page.screenshot({ path: 'verification.png', fullPage: true });
  });
});
