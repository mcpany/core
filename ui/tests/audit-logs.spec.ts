/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';

test.describe('Feature Screenshot', () => {
    const date = new Date().toISOString().split('T')[0];
    // Use test-results directory which is writable in CI
    const auditDir = path.join(process.cwd(), 'test-results/artifacts/audit/ui', date);

    test.beforeAll(async ({ playwright }) => {
        try {
            if (!fs.existsSync(auditDir)) {
                fs.mkdirSync(auditDir, { recursive: true });
            }
        } catch (e) {
            console.warn('Failed to create audit directory:', e);
        }

        // Seed database
        const apiContext = await playwright.request.newContext({
            baseURL: process.env.BACKEND_URL || 'http://localhost:50050',
            extraHTTPHeaders: {
                'X-API-Key': process.env.MCPANY_API_KEY || 'test-token',
            }
        });
        await apiContext.post('/api/v1/debug/seed', {
            data: {
                reset: true,
                services: [
                    {
                        name: "log-generator-service",
                        id: "log-generator-service",
                        httpService: { address: "http://example.com" }
                    }
                ]
            }
        });
    });

  test('Capture Logs', async ({ page }) => {
    await page.goto('/logs');
    // Wait for some logs to appear
    await page.waitForTimeout(3000);
    try {
        await page.screenshot({ path: path.join(auditDir, 'log_stream.png') });
    } catch (e) {
        console.warn('Failed to save screenshot:', e);
    }
  });
});
