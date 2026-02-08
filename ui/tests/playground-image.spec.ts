/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Rich Content', () => {
  const serviceName = 'image-test-service';

  test.afterEach(async ({ request }) => {
    // Cleanup
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});
  });

  test('should render image content from tool result', async ({ page, request }) => {
    // 1. Register command_line_service
    const registerRes = await request.post('/api/v1/services', {
      data: {
        id: serviceName,
        name: serviceName,
        command_line_service: {
          command: "echo",
          // working_directory: "/tmp", // Removed due to security policy
          tools: [
              {
                  name: "get_image",
                  description: "Returns an image",
                  call_id: "get_image_call",
              }
          ],
          calls: {
              "get_image_call": {
                  args: [
                      '{"content": [{"type": "image", "data": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=", "mimeType": "image/png"}]}'
                  ]
              }
          }
        }
      }
    });

    expect(registerRes.ok()).toBeTruthy();

    // 2. Navigate to Playground
    await page.goto('/playground');

    // 3. Select Tool
    await expect(page.getByText('get_image')).toBeVisible({ timeout: 15000 });
    await page.getByText('get_image').click();

    // 4. Run Tool
    await expect(page.getByRole('dialog')).toBeVisible();
    await page.getByRole('button', { name: /build command/i }).click();
    await page.getByLabel('Send').click();

    // 5. Verify Image
    // Check for errors first
    const errorLocator = page.locator('.text-destructive').first();
    if (await errorLocator.isVisible()) {
        console.error("Found error in chat:", await errorLocator.textContent());
    }

    try {
        await expect(page.getByRole('img', { name: /Result Image/i })).toBeVisible({ timeout: 15000 });
        const img = page.getByRole('img', { name: /Result Image/i }).first();
        await expect(img).toHaveAttribute('src', /data:image\/png;base64/);
    } catch (e) {
        // Fallback: check if we got raw JSON, which implies renderer issue
        const rawJsonLocator = page.getByText('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=');
        if (await rawJsonLocator.isVisible()) {
            console.error("Found raw JSON data, meaning renderer didn't pick it up.");
        }
        throw e;
    }
  });
});
