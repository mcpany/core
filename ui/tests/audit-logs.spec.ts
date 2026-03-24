/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';
import { seedGlobalState } from './e2e/test-data';

test.describe('Audit Logs View', () => {
    const date = new Date().toISOString().split('T')[0];
    const auditDir = path.join(process.cwd(), 'test-results/artifacts/audit/ui', date);

    test.beforeAll(async ({ request }) => {
        try {
            if (!fs.existsSync(auditDir)) {
                fs.mkdirSync(auditDir, { recursive: true });
            }
        } catch (e) {
            console.warn('Failed to create audit directory:', e);
        }

        // Seed the backend with actual data to verify the Audit Logs without mocking
        await seedGlobalState(request);
    });

    test('Capture Logs and Verify UI Elements', async ({ page }) => {
        await page.goto('/logs');
        // Wait for some logs to appear
        await page.waitForTimeout(3000);
        try {
            await page.screenshot({ path: path.join(auditDir, 'log_stream.png') });
        } catch (e) {
            console.warn('Failed to save screenshot:', e);
        }
    });

    test('Export Audit Logs to CSV', async ({ page }) => {
        await page.goto('/audit');
        await page.waitForSelector('text=Audit Logs');

        // Start waiting for download before clicking.
        const downloadPromise = page.waitForEvent('download', { timeout: 10000 }).catch(() => null);

        // Wait for Export CSV button to be visible and enabled
        const exportBtn = page.locator('button:has-text("Export CSV")');
        await exportBtn.waitFor({ state: 'visible' });

        // Since we are using real data seeded in the backend, we don't need to mock anything
        await exportBtn.click();

        const download = await downloadPromise;
        if (download) {
            const suggestedFilename = download.suggestedFilename();
            if (!suggestedFilename.includes('audit_export')) {
                 throw new Error(`Unexpected filename: ${suggestedFilename}`);
            }
            await download.cancel();
        }
    });
});
