/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Image Rendering', () => {
  const serviceName = 'image-test-service';

  test.beforeAll(async ({ request }) => {
    // Register the service
    // We use a command line service that echoes a JSON structure mimicking an MCP image result.
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        command_line_service: {
          command: 'echo',
          tools: [
            {
              name: 'generate_image',
              description: 'Generates a test image',
              call_id: 'generate_image'
            }
          ],
          calls: {
            'generate_image': {
              parameters: [],
              args: [
                 // JSON string for MCP content with image
                 JSON.stringify({
                     content: [
                         {
                             type: 'image',
                             data: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=',
                             mimeType: 'image/png'
                         }
                     ]
                 })
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

  test('should render image result from tool', async ({ page }) => {
    await page.goto('/playground');

    // Wait for tool to appear in sidebar
    await expect(page.getByText('generate_image')).toBeVisible();

    // Select the tool
    await page.getByText('generate_image').click();

    // Dialog should open
    await expect(page.getByRole('dialog')).toBeVisible();

    // Click Run (or Build Command)
    const runButton = page.getByRole('button', { name: /(run|execute|build)/i });
    await runButton.click();

    // Wait for dialog to close
    await expect(page.getByRole('dialog')).toBeHidden();

    // Click Send in the main input area
    await page.getByRole('button', { name: /send/i }).click();

    // Wait for result
    // The image should be visible
    const img = page.getByRole('img', { name: 'Tool Result' });
    await expect(img).toBeVisible();
    await expect(img).toHaveAttribute('src', /^data:image\/png;base64,/);
  });
});
