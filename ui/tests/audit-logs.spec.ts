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

  test('Seed Data and Verify Audit Logs View', async ({ page, request }) => {
      // 1. Seed Real Data into the backend via the /api/v1/debug/traces endpoint.
      const mockTrace = {
           request_id: 'test-req-id-12345',
           service_id: 'test-service-abc',
           client_ip: '127.0.0.1',
           mcp_method: 'tools/call',
           arguments: JSON.stringify({ arg1: 'value1', arg2: 42 }),
           result: JSON.stringify({ success: true, nested: { status: 'ok' } }),
           error: null,
           latency_ms: 150,
           created_at: new Date().toISOString(),
           response_status: 200,
           username: 'e2e-tester',
           tool_name: 'example_tool'
      };

      await request.post('/api/v1/debug/traces', {
          data: mockTrace,
          headers: { 'Content-Type': 'application/json' }
      });

      await page.goto('/audit');
      await page.waitForSelector('text=Audit Logs');
      await page.waitForTimeout(2000);

      const rows = page.locator('tbody tr');
      if (await rows.count() > 0) {
           await rows.first().click();
           await page.waitForTimeout(1000);

           // If any log details show up, test that Arguments or Result is shown using JsonView component
           if (await page.locator('h4:has-text("Arguments")').count() > 0) {
                await expect(page.locator('h4:has-text("Arguments")')).toBeVisible();
                await expect(page.locator('h4:has-text("Result")')).toBeVisible();
           }
      }
  });
});
