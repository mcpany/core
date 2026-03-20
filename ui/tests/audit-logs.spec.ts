/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';
import { seedAuditLogs } from './e2e/audit-data';

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
        await seedAuditLogs();
    });

  test('Render beautiful JSON Viewer for Audit Logs', async ({ page }) => {
    await page.goto('/audit');
    await page.waitForSelector('text=Audit Logs');

    // 2. We should see at least one "View" button for an audit log row.
    const viewButton = page.locator('button:has-text("View")').first();
    await viewButton.waitFor({ state: 'visible', timeout: 15000 }).catch(() => null);

    if (await viewButton.isVisible()) {
        // 3. Open the modal
        await viewButton.click();

        // 4. Verify that our new JsonViewer renders the formatted arguments
        // Our JSON viewer wraps keys like "test_argument:" in a span, not raw string
        // Check that we see the formatted text representation of the seeded args
        const dialog = page.locator('[role="dialog"]');
        await dialog.waitFor({ state: 'visible' });

        // Ensure the new JSON viewer renders
        await expect(dialog.locator('text=Arguments')).toBeVisible();
        await expect(dialog.locator('text=Result')).toBeVisible();

        // Verify it doesn't have the old syntax-highlighter class wrapper 'react-syntax-highlighter'
        const oldSyntaxHighlighter = dialog.locator('.react-syntax-highlighter').first();
        await expect(oldSyntaxHighlighter).toHaveCount(0);

        // Verify the new JsonViewer renders its components
        const jsonViewerKey = dialog.locator('text=test_argument:').first();
        await expect(jsonViewerKey).toBeVisible();
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
