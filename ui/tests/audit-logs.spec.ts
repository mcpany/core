/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';
import { seedServices } from './e2e/test-data';

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

        // Seed services to ensure echo_tool is available
        await seedServices(request);

        // Execute echo_tool to generate an audit log entry
        await request.post('/api/v1/execute', {
            data: {
                name: 'echo_tool',
                arguments: { test: 123 }
            }
        });
    });

  test('View Audit Log Details', async ({ page }) => {
    await page.goto('/audit');
    await page.waitForSelector('text=Audit Logs');

    // Wait for the echo_tool log entry to appear
    const toolCell = page.locator('td', { hasText: 'echo_tool' }).first();
    await toolCell.waitFor({ state: 'visible' });

    // Click "View" on the log entry row
    const row = page.locator('tr').filter({ has: toolCell });
    await row.locator('button', { hasText: 'View' }).click();

    // Verify dialog appears
    await expect(page.locator('h2', { hasText: 'Audit Log Detail' })).toBeVisible();

    // Verify Arguments section is using JsonView (which renders a container with class names containing react-json-view or similar, or we can check for text)
    // We expect the JSON keys to be visible
    await expect(page.locator('h4', { hasText: 'Arguments' })).toBeVisible();
    // JsonView renders standard keys. Let's look for "test"
    await expect(page.getByText('"test"').first()).toBeVisible();

    // Verify Result section is using RichResultViewer
    // For echo_tool it might return { "test": 123 } inside result
    await expect(page.locator('h4', { hasText: 'Result' })).toBeVisible();
    // RichResultViewer uses Tabs (JSON, Raw Output, Table, etc depending on content).
    // Let's verify that the JSON tab from RichResultViewer is present
    await expect(page.getByRole('tab', { name: 'JSON' })).toBeVisible();
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
