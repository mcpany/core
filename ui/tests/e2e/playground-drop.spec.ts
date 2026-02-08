/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';

test.describe('Playground Smart File Drop', () => {
    // Seed test data
    test.beforeAll(async ({ request }) => {
        // Seed service with image tool
        const res = await request.post('/api/v1/services', {
            data: {
                id: "svc_img",
                name: "ImageProcessor",
                version: "1.0",
                http_service: {
                    address: "http://localhost:9090",
                    tools: [
                        {
                            name: "analyze_image",
                            description: "Analyze an image",
                            inputSchema: {
                                type: "object",
                                properties: {
                                    image: {
                                        type: "string",
                                        contentEncoding: "base64",
                                        contentMediaType: "image/png"
                                    }
                                }
                            }
                        }
                    ]
                }
            },
            headers: { 'X-API-Key': 'test-token' } // Use consistent API key if possible, or assume none for dev
        });
        expect(res.ok()).toBeTruthy();
    });

    test.afterAll(async ({ request }) => {
        try {
            await request.delete('/api/v1/services/ImageProcessor', {
                headers: { 'X-API-Key': 'test-token' }
            });
        } catch (e) {
            console.error('Failed to cleanup test service:', e);
        }
    });

    test('should detect compatible tool on file drop and show preview', async ({ page }) => {
        await page.goto('/playground');

        // Create a fake image file
        // A minimal 1x1 PNG pixel
        const buffer = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==', 'base64');

        // Define data transfer
        const dataTransfer = await page.evaluateHandle((data) => {
            const dt = new DataTransfer();
            // Convert base64 back to Uint8Array inside browser context
            const binaryString = atob(data.base64);
            const bytes = new Uint8Array(binaryString.length);
            for (let i = 0; i < binaryString.length; i++) {
                bytes[i] = binaryString.charCodeAt(i);
            }
            const file = new File([bytes], 'test-image.png', { type: 'image/png' });
            dt.items.add(file);
            return dt;
        }, { base64: buffer.toString('base64') });

        // Dispatch drop event on the drop zone
        await page.dispatchEvent('[data-testid="smart-drop-zone"]', 'drop', { dataTransfer });

        // Wait for ToolForm dialog to open
        // Since we have only 1 tool (analyze_image) compatible with image/png, it should auto-select.
        await expect(page.getByRole('dialog')).toBeVisible({ timeout: 10000 });

        // Verify correct tool is selected
        await expect(page.getByText('analyze_image')).toBeVisible();

        // Verify file input shows file name
        await expect(page.getByText('test-image.png')).toBeVisible();

        // Verify preview image exists (enhanced FileInput)
        // The enhanced FileInput renders an `img` tag if previewUrl is set.
        const img = page.locator('img[alt="Preview"]');
        await expect(img).toBeVisible();

        // Verify source is a blob URL
        const src = await img.getAttribute('src');
        expect(src).toMatch(/^blob:/);
    });
});
