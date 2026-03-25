/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';
import { seedUser, seedGlobalState } from './e2e/test-data';

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

  test('Export Audit Logs to CSV', async ({ page, request }) => {
    await seedGlobalState(request);
    await seedUser(request, "e2e-admin-core");

    // Login
    await page.goto('/login');
    await page.fill('input[name="username"]', 'e2e-admin-core');
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/');

    await page.goto('/audit');
    await page.waitForSelector('text=Audit Logs');

    // Start waiting for download before clicking.
    const downloadPromise = page.waitForEvent('download', { timeout: 10000 }).catch(() => null);

    // Wait for Export CSV button to be visible and enabled
    const exportBtn = page.locator('button:has-text("Export CSV")');
    await exportBtn.waitFor({ state: 'visible' });

    // The backend handles /api/v1/audit/export naturally since we seeded it.
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

  test('Render Rich JSON Tables for Audit Logs', async ({ page, request }) => {
    await seedGlobalState(request);
    await seedUser(request, "e2e-admin-core");

    // Login
    await page.goto('/login');
    await page.fill('input[name="username"]', 'e2e-admin-core');
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/');

    await page.goto('/audit');

    // Wait for the table to load
    await page.waitForSelector('table');

    // Find the row for process_payment
    const row = page.locator('tr', { hasText: 'process_payment' }).first();
    await expect(row).toBeVisible();

    // Click the "View" button to open the dialog
    await row.locator('button:has-text("View")').click();

    // Wait for the dialog
    const dialog = page.locator('div[role="dialog"]');
    await expect(dialog).toBeVisible();

    // Expect the dialog to display the JSON as a rich table
    // Look for the "Table" tab in the arguments or result rich viewer
    const tableTab = dialog.locator('button[role="tab"]', { hasText: 'Table' }).first();
    await expect(tableTab).toBeVisible();
    await tableTab.click();

    // Now look for the table cells in the result rich viewer
    await expect(dialog.locator('td', { hasText: 'ch_123' }).first()).toBeVisible();
    await expect(dialog.locator('td', { hasText: 'succeeded' }).first()).toBeVisible();
  });
});
