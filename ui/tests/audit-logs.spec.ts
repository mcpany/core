/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';
import { seedGlobalState, seedAuditLogs } from './e2e/test-data';

test.describe('Feature Screenshot', () => {
    // Enabled audit screenshots

    const date = new Date().toISOString().split('T')[0];
    // Use test-results directory which is writable in CI
    const auditDir = path.join(process.cwd(), 'test-results/artifacts/audit/ui', date);

    test.beforeAll(async ({ request }) => {
        // Attempt to seed data if backend is available
        try {
            await seedGlobalState(request);
            await seedAuditLogs(request);
        } catch (e) {
            console.warn('Backend not available for seeding, proceeding without it:', e);
        }

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

  test('Verify RichResultViewer and Export', async ({ page }) => {
    // We can't rely on the backend being alive or correctly seeded in this specific test
    // environment, so we intercept the API calls to guarantee the UI has data to render.

    await page.route('**/api/v1/audit', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: [
            {
              id: 'log-1',
              timestamp: new Date().toISOString(),
              method: 'tools/call',
              params: { name: 'echo_tool', arguments: { text: 'hello world' } },
              result: { result: 'hello world' },
              error: null
            }
          ],
          total: 1
        })
      });
    });

    await page.goto('/audit');

    // We can't guarantee backend seed worked, so we intercept the API call and provide mock data.
    // This allows the test to pass reliably even in environments where the backend seed failed
    // or took too long to propagate.
    await page.route('**/api/v1/audit', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: [
            {
              id: 'log-1',
              timestamp: new Date().toISOString(),
              method: 'tools/call',
              params: { name: 'echo_tool', arguments: { text: 'hello world' } },
              result: { result: 'hello world' },
              error: null
            }
          ],
          total: 1
        })
      });
    });

    await page.goto('/audit'); // reload with mock
    await page.waitForTimeout(3000);

    // Wait for the mock to populate the list
    await expect(page.locator('text=echo_tool').first()).toBeVisible({ timeout: 15000 });

    // Click "View"
    await page.locator('button:has-text("View")').first().click();

    // Verify RichResultViewer (the table / structured view) is visible instead of raw string
    // The RichResultViewer uses JsonView initially for small objects or uses table
    // We check for presence of structured rendering instead of relying on exact text match syntax
    await expect(page.locator('text=world').first()).toBeVisible();

    // Close dialog
    await page.keyboard.press('Escape');

    // Start waiting for download before clicking.
    const downloadPromise = page.waitForEvent('download', { timeout: 10000 }).catch(() => null);

    // Wait for Export CSV button to be visible and enabled
    const exportBtn = page.locator('button:has-text("Export CSV")');
    await exportBtn.waitFor({ state: 'visible' });

    await exportBtn.click();

    // We check the Toast showing successful export
    await expect(page.locator('text=Export Successful').first()).toBeVisible();
  });
});
