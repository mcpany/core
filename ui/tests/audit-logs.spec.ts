/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';

test.describe('Feature Screenshot', () => {
    // Enabled audit screenshots unconditionally for full coverage
    // test.skip(process.env.CAPTURE_SCREENSHOTS !== 'true', 'Skipping audit screenshots');

    const date = new Date().toISOString().split('T')[0];
    // Use test-results directory which is writable in CI
    const auditDir = path.join(process.cwd(), 'test-results/artifacts/audit/ui', date);

    test.beforeAll(async () => {
        try {
            if (!fs.existsSync(auditDir)) {
                fs.mkdirSync(auditDir, { recursive: true });
            }
        } catch (e) {
            console.warn('Failed to create audit directory:', e);
        }
    });

  test('Capture Logs', async ({ page, request }) => {
    // Generate some logs first
    // We can call the health endpoint which is often logged, or any API
    // Or assume the server has logs from startup.
    // To be safe, let's call a tool or API if possible, but we don't have tool definitions handy.
    // Calling /api/v1/system/status usually works.
    await request.get('/api/v1/system/status');

    await page.goto('/logs');
    // Wait for the page to load
    await expect(page.locator('h1')).toContainText('Logs');

    // Wait for some logs to appear or at least the page to load
    // We wait for the table or empty state
    try {
        await page.waitForSelector('table, div:has-text("No logs found")', { timeout: 5000 });
    } catch (e) {
        // Ignore timeout, we just want to ensure page is roughly ready
    }

    await page.waitForTimeout(1000);
    try {
        await page.screenshot({ path: path.join(auditDir, 'log_stream.png') });
    } catch (e) {
        console.warn('Failed to save screenshot:', e);
    }
  });
});
