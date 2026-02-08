/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Image Rendering', () => {
    const serviceName = 'image-test-service-' + Date.now();

    test.beforeAll(async ({ request }) => {
        // Register a command line service that returns an image
        // We use 'echo' to simulate a tool returning a JSON string that conforms to MCP CallToolResult structure
        const response = await request.post('/api/v1/services', {
            data: {
                id: serviceName,
                name: serviceName,
                command_line_service: {
                    command: "echo",
                    // We define a tool 'get_image' that maps to a call 'call_1'
                    tools: [
                        {
                            name: "get_image",
                            description: "Returns a 1x1 red pixel image",
                            call_id: "call_1",
                            input_schema: { type: "object", properties: {} }
                        }
                    ],
                    // We define the call 'call_1' which passes the JSON string as an argument to 'echo'
                    calls: {
                        "call_1": {
                            id: "call_1",
                            args: [
                                '{"content": [{"type": "image", "data": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=", "mimeType": "image/png"}]}'
                            ]
                        }
                    }
                }
            }
        });

        // Log error if registration fails
        if (!response.ok()) {
            console.error('Failed to register service:', await response.text());
        }
        expect(response.ok()).toBeTruthy();
    });

    test.afterAll(async ({ request }) => {
        // Cleanup
        await request.delete(`/api/v1/services/${serviceName}`);
    });

    test('should render image result from real tool execution', async ({ page }) => {
        await page.goto('/playground');

        // Search and select the tool
        // Focus to avoid hydration issues if any
        await page.getByPlaceholder('Enter command or select a tool...').click();

        // Search in the sidebar
        const sidebarSearch = page.getByPlaceholder('Search tools...');
        if (await sidebarSearch.isVisible()) {
            await sidebarSearch.fill('get_image');
        }

        // Wait for tool to appear in the list
        await expect(page.getByText('get_image')).toBeVisible();

        // Click the tool to open configuration dialog
        await page.getByText('get_image').click();

        // Expect dialog to open
        await expect(page.getByRole('dialog')).toBeVisible();
        await expect(page.getByRole('heading', { name: 'get_image' })).toBeVisible();

        // Click "Build Command"
        await page.getByRole('button', { name: /build command/i }).click();

        // Dialog should close
        await expect(page.getByRole('dialog')).toBeHidden();

        // Input should be populated with 'get_image {}'
        await expect(page.getByRole('textbox', { name: /enter command/i })).toHaveValue(/get_image/);

        // Click "Send" button
        await page.getByRole('button', { name: 'Send' }).click();

        // Wait for result to appear. The result will contain an image.
        // We look for an img tag with the specific data URI prefix.
        // The base64 data starts with "iVBOR..."
        const imgSelector = 'img[src^="data:image/png;base64,iVBOR"]';

        // Increase timeout as tool execution (CLI) might take a moment
        await expect(page.locator(imgSelector)).toBeVisible({ timeout: 10000 });

        // Also verify the "Image (image/png)" label exists
        await expect(page.getByText('Image (image/png)')).toBeVisible();
    });
});
