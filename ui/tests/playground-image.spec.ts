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
                                JSON.stringify([
                                    {
                                        type: 'image',
                                        data: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=',
                                        mimeType: 'image/png'
                                    }
                                ])
                            ]
                        }
                    }
                }
            }
        });
        expect(response.ok()).toBeTruthy();
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

        await expect(page.getByRole('dialog')).toBeVisible();
        await expect(page.getByText('Configure arguments')).toBeVisible();

        // Since no args required, we can just click "Build Command"
        // The dialog has a form.
        // There should be a submit button.
        await page.getByRole('button', { name: 'Build Command' }).click();

        // Dialog closes, command is populated in input
        await expect(page.getByRole('dialog')).toBeHidden();

        // Click "Send" button to execute
        await page.getByLabel('Send').click();

        // Wait for result
        // Result should contain an image
        // We look for img tag with the specific src prefix
        const srcPrefix = 'data:image/png;base64,iVBOR';

        // Wait for message to appear
        await expect(page.locator(`img[src^="${srcPrefix}"]`)).toBeVisible({ timeout: 10000 });

        // Also check if "Rich" view button is active/visible
        // Since logic forces Rich view for images, it should be rendered directly.
        // We can check if "Rich" button exists in the result renderer toolbar.
        await expect(page.getByRole('button', { name: 'Rich' })).toBeVisible();
    });
});
