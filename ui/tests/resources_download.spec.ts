/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Download Feature', () => {
    test.beforeEach(async ({ page }) => {
        // Mock resources list
        await page.route('**/api/v1/resources', async (route) => {
            await route.fulfill({
                json: {
                    resources: [
                        {
                            name: 'test-file.txt',
                            uri: 'file://test-file.txt',
                            mimeType: 'text/plain'
                        }
                    ]
                }
            });
        });

        // Mock resource content read (for the preview panel AND download)
        await page.route('**/api/v1/resources/read*', async (route) => {
             const urlObj = new URL(route.request().url());
             const uri = urlObj.searchParams.get('uri');

             if (uri === 'file://test-file.txt') {
                 await route.fulfill({
                     json: {
                         contents: [{
                             uri: 'file://test-file.txt',
                             mimeType: 'text/plain',
                             blob: Buffer.from('Hello World Content').toString('base64') // Use blob to test decoding
                         }]
                     }
                 });
             } else {
                 await route.fulfill({ status: 404 });
             }
        });
    });

    test('should trigger download via client-side blob', async ({ page }) => {
        await page.goto('/resources');

        // Wait for list to load
        await page.waitForSelector('div.font-medium.truncate');

        // Select the resource to enable download button
        await page.locator('div.font-medium.truncate', { hasText: 'test-file.txt' }).first().click();

        const downloadPromise = page.waitForEvent('download');

        // Click the download button in the toolbar
        await page.getByRole('button', { name: 'Download' }).click();

        const download = await downloadPromise;
        expect(download.suggestedFilename()).toBe('test-file.txt');

        // We can optionally verify the content stream
        const stream = await download.createReadStream();
        // ... (reading stream is async and might require stream consumers)
    });

    test('Drag start should set DownloadURL correctly', async ({ page }) => {
        await page.goto('/resources');
        await page.waitForSelector('div[draggable="true"]');

        const dragData: any = await page.evaluate(async () => {
            return new Promise((resolve) => {
                const item = document.querySelector('div[draggable="true"]');

                if (!item) {
                    resolve({ error: 'Item not found' });
                    return;
                }

                const dataTransfer = new DataTransfer();
                const originalSetData = dataTransfer.setData.bind(dataTransfer);
                const capturedData: Record<string, string> = {};

                dataTransfer.setData = (format, data) => {
                    capturedData[format] = data;
                    return originalSetData(format, data);
                };

                const event = new DragEvent('dragstart', {
                    bubbles: true,
                    cancelable: true,
                    dataTransfer: dataTransfer
                });

                item.dispatchEvent(event);

                setTimeout(() => {
                   resolve(capturedData);
                }, 100);
            });
        });

        if (dragData.error) {
            throw new Error(dragData.error);
        }

        expect(dragData['text/plain']).toBe('file://test-file.txt');
        expect(dragData['DownloadURL']).toBeDefined();

        const parts = dragData['DownloadURL'].split(':');
        expect(parts[0]).toBe('text/plain');
        expect(parts[1]).toBe('test-file.txt');
        const url = parts.slice(2).join(':');
        expect(url).toContain('/api/resources/download?uri=file%3A%2F%2Ftest-file.txt&name=test-file.txt');
    });
});
