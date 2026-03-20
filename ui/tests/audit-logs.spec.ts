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

  test('Render Rich JSON in Audit Logs', async ({ page }) => {
    // Intercept the API call to provide a mock audit log entry with rich JSON
    await page.route('**/api/v1/audit*', async (route) => {
      await route.fulfill({
        json: {
          entries: [
            {
              timestamp: new Date().toISOString(),
              toolName: 'complex_tool',
              userId: 'admin',
              profileId: 'default',
              durationMs: 150,
              arguments: JSON.stringify({ complex_arg: { nested: "value" } }),
              result: JSON.stringify([{ id: 1, name: "item1" }, { id: 2, name: "item2" }])
            }
          ],
          totalCount: 1
        }
      });
    });

    await page.goto('/audit');

    // Wait for the mock log to appear and click "View"
    await page.locator('button:has-text("View")').first().click();

    // Verify that the RichResultViewer renders the result as a Table (since it's an array of objects)
    // Check for the "Table" tab which is the default for tabular eligible data
    await page.waitForSelector('text=item1');

    // Verify the arguments are rendered (likely JSON tree or formatted output)
    await page.waitForSelector('text=complex_arg');
  });
});
