/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState, seedTraces } from './test-data';

test.describe('Audit Logs Viewer', () => {
  test.beforeEach(async ({ page, request }) => {
    // We reuse the seedGlobalState and the trace seeder to populate the audit logs db
    // The handleDebugSeedTraces in api_traces.go inserts into the audit DB
    try {
      await seedGlobalState(request);
      await page.waitForTimeout(2000);
      await seedTraces(request);
      await page.waitForTimeout(2000);
    } catch (e) {
      console.log(`Failed to seed: ${e}`);
    }

    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', 'e2e-admin-core');
    await page.fill('input[name="password"]', 'password');
    await Promise.all([
      page.waitForURL('/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
    ]);
  });

  test('should display split-pane layout and SmartTable for JSON array output', async ({ page }) => {
    // Audit logs viewer gets data via websocket occasionally but listAuditLogs is also called
    await page.goto('/audit');

    // Give it plenty of time for websocket/network data to propagate in CI
    await page.waitForTimeout(5000);

    const row = page.locator('text=list-users').first();
    await expect(row).toBeVisible({ timeout: 20000 });

    await row.click();

    await expect(page.locator('text=Audit Log Detail')).toBeVisible();

    await expect(page.getByPlaceholder('Search all columns...')).toBeVisible();

    await expect(page.locator('text=Alice').first()).toBeVisible();
    await expect(page.locator('text=admin').first()).toBeVisible();
  });
});
