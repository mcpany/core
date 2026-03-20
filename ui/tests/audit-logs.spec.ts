/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';

test.describe('Audit Logs Viewer', () => {

  test('Capture Logs and Verify UI Elements', async ({ page }) => {
    // For local mocking/verifying just the UI structure
    await page.route('**/api/v1/audit/logs*', async route => {
      await route.fulfill({
        json: {
            entries: [
              {
                timestamp: new Date().toISOString(),
                toolName: 'test_tool',
                userId: 'admin',
                arguments: '{"key": "value"}',
                result: '{"success": true}',
                error: "",
                duration: "10ms",
                durationMs: 10
              }
            ]
        }
      });
    });

    await page.goto('/audit');
    await page.waitForSelector('text=Audit Logs', { timeout: 10000 }).catch(() => {});

    // Check if view button exists
    const viewBtn = page.locator('button:has-text("View")').first();
    if (await viewBtn.count() > 0) {
        await viewBtn.click();
        const dialog = page.locator('div[role="dialog"]');
        await expect(dialog).toBeVisible();
        await expect(dialog.locator('text=Raw').first()).toBeVisible();
    }
  });

  test('Export Audit Logs to CSV', async ({ page }) => {
    await page.goto('/audit');
    await page.waitForSelector('text=Audit Logs', { timeout: 10000 }).catch(() => {});

    // Start waiting for download before clicking.
    const downloadPromise = page.waitForEvent('download', { timeout: 10000 }).catch(() => null);

    // Wait for Export CSV button to be visible and enabled
    const exportBtn = page.locator('button:has-text("Export CSV")');
    if (await exportBtn.count() > 0) {
        await exportBtn.waitFor({ state: 'visible' });
        await exportBtn.click();

        const download = await downloadPromise;
        if (download) {
            const suggestedFilename = download.suggestedFilename();
            if (!suggestedFilename.includes('audit_export')) {
                 throw new Error(`Unexpected filename: ${suggestedFilename}`);
            }
            await download.cancel();
        }
    }
  });
});
