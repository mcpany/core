/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';

test.describe('Feature Screenshot', () => {
    // Audit screenshots should be enabled for CI/E2E runs

    const date = new Date().toISOString().split('T')[0];
    const auditDir = path.join(__dirname, '../.audit/ui', date);

    test.beforeAll(async () => {
        if (!fs.existsSync(auditDir)) {
            fs.mkdirSync(auditDir, { recursive: true });
        }
    });

  test('Capture Logs', async ({ page }) => {
    await page.goto('/logs');
    // Wait for some logs to appear
    await page.waitForTimeout(3000);
    // If CAPTURE_SCREENSHOTS is true, take screenshot
    if (process.env.CAPTURE_SCREENSHOTS === 'true') {
        await page.screenshot({ path: path.join(auditDir, 'log_stream.png') });
    }
  });
});
