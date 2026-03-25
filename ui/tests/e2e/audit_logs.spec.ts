/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Audit Logs Viewer', () => {
  test.beforeEach(async ({ page, request }) => {
    // We reuse the seedGlobalState and the trace seeder to populate the audit logs db
    // The handleDebugSeedTraces in api_traces.go inserts into the audit DB
    await seedGlobalState(request);
    await request.post('/api/v1/debug/traces', { data: {}, headers: { 'Authorization': 'Bearer test-token' } });

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
    await page.goto('/audit');

    await page.waitForSelector('text=list-users', { state: 'visible', timeout: 10000 });

    await page.click('text=list-users');

    await expect(page.locator('text=Audit Log Detail')).toBeVisible();

    await expect(page.getByPlaceholder('Search all columns...')).toBeVisible();

    await expect(page.locator('text=Alice').first()).toBeVisible();
    await expect(page.locator('text=admin').first()).toBeVisible();
  });
});
