/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Image Rendering', () => {
    const SERVICE_NAME = 'image-test-service';

    test.beforeAll(async ({ request }) => {
        // cleanup
        try {
            await request.delete(`/api/v1/services/${SERVICE_NAME}`);
        } catch (_) {}

        // Register a command line service
        const response = await request.post('/api/v1/services', {
            data: {
                name: SERVICE_NAME,
                command_line_service: {
                    command: 'echo',
                    tools: [
                        { name: 'get_image', call_id: 'call1', description: 'Returns an image' }
                    ],
                    calls: {
                        'call1': {
                            args: [
                                JSON.stringify({
                                    content: [{
                                        type: 'image',
                                        data: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=',
                                        mimeType: 'image/png'
                                    }]
                                })
                            ]
                        }
                    }
                }
            }
        });
        if (!response.ok() && response.status() !== 409) {
            console.error(`Status: ${response.status()}, Body: ${await response.text()}`);
        }
        expect(response.ok() || response.status() === 409).toBeTruthy();
    });

    test.afterAll(async ({ request }) => {
        try {
            await request.delete(`/api/v1/services/${SERVICE_NAME}`);
        } catch (_) {}
    });

    test('should render image tool result in playground', async ({ page }) => {
        await page.goto('/playground');

        // Wait for services to load and click on the service
        // The sidebar loads tools. We can search for the tool.
        // Assuming the tool appears in the sidebar.

        // Wait for sidebar to populate
        await expect(page.getByText('get_image')).toBeVisible({ timeout: 10000 });

        // Click the tool "get_image"
        await page.getByText('get_image').click();

        // Modal opens? Or it fills the input?
        // Playground Pro behaviour:
        // clicking tool in sidebar opens "Configure arguments" dialog.

        await expect(page.getByRole('tab', { name: 'Tool Runner' })).toBeVisible();

        // Since no args required, we can just click "Execute"
        await page.getByRole('button', { name: 'Execute', exact: true }).click();

        // No need to click Send, it auto executes

        // Wait for result
        // Result should contain an image
        // We look for img tag with the specific src prefix
        const srcPrefix = 'data:image/png;base64,iVBOR';

        // Wait for message to appear
        // Increased timeout for CI stability
        try {
            await expect(page.locator('img[src^="data:image/png"]')).toBeVisible({ timeout: 60000 });
        } catch (e) {
            const imgs = await page.evaluate(() => Array.from(document.querySelectorAll('img')).map(i => i.src.substring(0, 50)));
            console.log('Found images:', imgs);
            throw e;
        }

        // Rich view is verified by the image presence.
    });
});
