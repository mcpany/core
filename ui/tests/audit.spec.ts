/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import * as path from 'path';

test.describe('Audit Screenshots', () => {
  const date = new Date().toISOString().split('T')[0];
  const auditDir = path.join(__dirname, '../../.audit/ui', date);

  test('Capture Dashboard', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(auditDir, 'dashboard.png'), fullPage: true });
  });

  test('Capture Services', async ({ page }) => {
    await page.goto('/services');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(auditDir, 'services.png'), fullPage: true });
  });

  test('Capture Tools', async ({ page }) => {
    await page.goto('/tools');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(auditDir, 'tools.png'), fullPage: true });
  });

   test('Capture Middleware', async ({ page }) => {
    await page.goto('/middleware');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(auditDir, 'middleware.png'), fullPage: true });
  });

  test('Capture Settings', async ({ page }) => {
    await page.goto('/settings');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(auditDir, 'settings.png'), fullPage: true });
  });
});
