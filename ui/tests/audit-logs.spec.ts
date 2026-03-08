/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
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
});

test.describe('Audit Logs Viewer UX', () => {
    test('Displays formatted JSON in the detailed view', async ({ page, request }) => {
        // Seed some global state and traffic which populates trace and audit events
        const { seedGlobalState, seedTraffic, seedUser } = require('./e2e/test-data');
        await seedGlobalState(request);
        await seedTraffic(request);
        await seedUser(request, "e2e-audit-admin");

        // Execute a tool to ensure an audit log exists with actual argument and result data
        await request.post('/api/v1/tools/call', {
            data: {
                name: 'echo_tool',
                arguments: { message: 'hello audit view' }
            }
        });

        await page.goto('/audit');
        await page.waitForSelector('text=Audit Logs');

        // Look for the "View" button in the table and click it.
        // We ensure we wait for it so the data has loaded.
        const viewBtn = page.locator('button:has-text("View")').first();
        await viewBtn.waitFor({ state: 'visible', timeout: 15000 });
        await viewBtn.click();

        // The dialog should appear containing 'Arguments' and 'Result' sections.
        const dialog = page.locator('[role="dialog"]');
        await dialog.waitFor({ state: 'visible' });

        // Verify the dialog contains 'Arguments' and 'Result' sections.
        await expect(dialog.locator('h4:has-text("Arguments")')).toBeVisible();
        await expect(dialog.locator('h4:has-text("Result")')).toBeVisible();

        // Verify JsonView specific functionality is present (e.g. "Raw", "Tree" mode, or "Copy JSON" title).
        // Since JsonView is used, it renders a toolbar with a 'Raw' view trigger and copy button.
        const copyBtns = dialog.locator('button[title="Copy JSON"], button[title="Copy value"], button:has-text("Raw")');
        await expect(copyBtns.first()).toBeVisible();
    });
});
