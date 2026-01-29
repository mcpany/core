/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import http from 'http';
import { AddressInfo } from 'net';

const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:50050';
// Using a simpler API key/Token strategy if needed, but defaults usually work in dev
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';

test.describe('Audit Logs Rich Content', () => {
    let mockServer: http.Server;
    let mockUrl: string;
    const SERVICE_NAME = 'audit-test-rich';

    test.beforeAll(async () => {
        // Start a mock server that returns the rich content
        mockServer = http.createServer((req, res) => {
            if (req.url === '/image' && req.method === 'POST') {
                res.writeHead(200, { 'Content-Type': 'application/json' });
                res.end(JSON.stringify({
                    content: [
                        {
                            type: "image",
                            data: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
                            mimeType: "image/png"
                        },
                        {
                             type: "text",
                             text: "Rich text content"
                        }
                    ]
                }));
            } else {
                res.writeHead(404);
                res.end();
            }
        });

        await new Promise<void>((resolve) => {
            mockServer.listen(0, '127.0.0.1', () => resolve());
        });
        const port = (mockServer.address() as AddressInfo).port;
        mockUrl = `http://127.0.0.1:${port}`;
        console.log(`Mock server listening at ${mockUrl}`);
    });

    test.afterAll(async ({ request }) => {
        mockServer.close();
        // Cleanup service
         await request.delete(`${BACKEND_URL}/api/v1/services/${SERVICE_NAME}`, {
             headers: { 'X-API-Key': API_KEY }
        }).catch(() => {});
    });

    test('should display rich content in audit logs', async ({ page, request }) => {
        // 1. Register Service
        // Ensure cleanup first
        await request.delete(`${BACKEND_URL}/api/v1/services/${SERVICE_NAME}`, {
             headers: { 'X-API-Key': API_KEY }
        }).catch(() => {});

        const registerRes = await request.post(`${BACKEND_URL}/api/v1/services`, {
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': API_KEY
            },
            data: {
                id: SERVICE_NAME,
                name: SERVICE_NAME,
                version: "1.0.0",
                http_service: {
                    address: mockUrl,
                    tools: [
                        { name: "generate_image", description: "Returns an image" }
                    ],
                    calls: {
                        "generate_image": {
                            id: "generate_image",
                            method: 2, // POST
                            endpoint_path: "/image"
                        }
                    }
                }
            }
        });
        expect(registerRes.ok(), `Failed to register service: ${registerRes.statusText()}`).toBeTruthy();

        // 2. Execute Tool to generate log
        const execRes = await request.post(`${BACKEND_URL}/api/v1/execute`, {
             headers: {
                'Content-Type': 'application/json',
                'X-API-Key': API_KEY
            },
            data: {
                name: "generate_image",
                arguments: {}
            }
        });
        expect(execRes.ok(), `Failed to execute tool: ${execRes.statusText()}`).toBeTruthy();

        // 3. Navigate to Audit Log
        await page.goto('/audit');

        // 4. Find the log
        // Fill filter
        await page.getByPlaceholder('e.g. weather_get').fill('generate_image');
        await page.getByRole('button', { name: 'Filter' }).click();

        // Wait for table to reload
        await page.waitForTimeout(1000);

        // Click View - assume the top one is ours or the only one
        const viewBtn = page.getByRole('button', { name: 'View' }).first();
        await expect(viewBtn).toBeVisible();
        await viewBtn.click();

        // 5. Verify Content
        // Check for Image
        const img = page.locator('img[alt^="Result"]');
        await expect(img).toBeVisible();
        await expect(img).toHaveAttribute('src', /data:image\/png;base64/);

        // Check for Text
        await expect(page.getByText('Rich text content')).toBeVisible();

        // 6. Verify Fallback Tab
        await page.getByRole('tab', { name: 'Raw Data' }).click();
        await expect(page.getByText('"mimeType": "image/png"')).toBeVisible();
    });
});
