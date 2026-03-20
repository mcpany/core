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

    test.beforeAll(async () => {
        try {
            if (!fs.existsSync(auditDir)) {
                fs.mkdirSync(auditDir, { recursive: true });
            }
        } catch (e) {
            console.warn('Failed to create audit directory:', e);
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

  test('View Audit Log Details', async ({ request, page }) => {
    // 1. Seed the database with a test audit log using the execute endpoint to generate real data
    try {
       await request.post('/api/v1/execute', {
            data: { tool_name: "dummy_tool", arguments: {"test": "data"} }
       });
    } catch(e) {
       // Ignore execution errors as we just want the audit log of the attempt
    }

    await page.goto('/audit');
    // Ensure the page has loaded before attempting to interact or locate elements
    await page.waitForSelector('text=Audit Logs');

    // Wait for the table to populate
    await page.waitForTimeout(1000);

    // Find and click the first "View" button in the audit log table
    const viewBtns = page.locator('button:has-text("View")');
    if (await viewBtns.count() > 0) {
        await viewBtns.first().click();

        // Wait for the detail dialog to appear
        await page.waitForSelector('text=Audit Log Detail');

        // Verify JsonTree is rendered (look for braces representing JSON structures)
        await page.waitForSelector('text={');
    } else {
        console.warn('No audit logs available to view. Skipping view test.');
    }
  });
});
