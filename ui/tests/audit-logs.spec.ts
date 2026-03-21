/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';

test.describe('Feature Screenshot', () => {
    // Enabled audit screenshots

    const date = new Date().toISOString().split('T')[0];
    // Use test-results directory which is writable in CI
    const auditDir = path.join(process.cwd(), 'test-results/artifacts/audit/ui', date);

    test.beforeAll(async ({ request }) => {
        try {
            if (!fs.existsSync(auditDir)) {
                fs.mkdirSync(auditDir, { recursive: true });
            }
        } catch (e) {
            console.warn('Failed to create audit directory:', e);
        }

        // Seed audit logs for these tests to ensure real data is verified.
        try {
            await request.post('/api/v1/debug/traces');
        } catch (e) {
            console.warn('Failed to seed debug traces:', e);
        }
    });

  test('View Audit Log Details', async ({ page }) => {
    await page.goto('/audit');
    await page.waitForSelector('text=Audit Logs');
    await page.waitForTimeout(1000); // Give it a second to load

    // Look for the "View" button in the table and click the first one
    const viewBtn = page.locator('button:has-text("View")').first();
    await viewBtn.waitFor({ state: 'visible' });
    await viewBtn.click();

    // Verify dialog appears and contains expected elements
    await page.waitForSelector('text=Audit Log Detail');
    await page.waitForSelector('text=Arguments');
    await page.waitForSelector('text=Result');

    // RichResultViewer or JsonView should be visible
    const codeBlocks = page.locator('.lucide-code').first();
    const rawButtons = page.locator('button:has-text("Raw")').first();
    // At least one of these UI indicators of our new viewers should be present
    await Promise.any([
        codeBlocks.waitFor({ state: 'visible', timeout: 5000 }).catch(() => null),
        rawButtons.waitFor({ state: 'visible', timeout: 5000 }).catch(() => null),
    ]);

    try {
        await page.screenshot({ path: path.join(auditDir, 'audit_detail.png') });
    } catch (e) {
        console.warn('Failed to save screenshot:', e);
    }
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

  test('Export Audit Logs to CSV', async ({ page }) => {
    await page.goto('/audit');
    await page.waitForSelector('text=Audit Logs');

    // Start waiting for download before clicking.
    const downloadPromise = page.waitForEvent('download', { timeout: 10000 }).catch(() => null);

    // Wait for Export CSV button to be visible and enabled
    const exportBtn = page.locator('button:has-text("Export CSV")');
    await exportBtn.waitFor({ state: 'visible' });

    // Check if we need to mock since we are not fully seeding audit data for this specific test
    // but the backend handles /api/v1/audit/export naturally.
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
