/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test } from '@playwright/test';
import * as path from 'path';

test.describe('Feature Screenshot', () => {
  const date = new Date().toISOString().split('T')[0];
  const auditDir = path.join(__dirname, '../.audit/ui', date);

  test('Capture Logs', async ({ page }) => {
    await page.goto('/logs');
    // Wait for some logs to appear
    await page.waitForTimeout(3000);
    await page.screenshot({ path: path.join(auditDir, 'log_stream.png') });
  });
});
