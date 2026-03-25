/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Audit Logs Viewer', () => {
  test.beforeEach(async ({ page, request }) => {
    try {
      await seedGlobalState(request);
      await page.waitForTimeout(1000);

      const res = await request.post('/api/v1/debug/traces', { data: {}, headers: { 'Authorization': 'Bearer test-token', 'Content-Type': 'application/json' } });
      console.log('Seeding status:', res.status());

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

  test('should display split-pane layout and SmartTable for JSON array output', async ({ page, request }) => {
    // Audit logs viewer gets data via websocket occasionally but listAuditLogs is also called
    await page.goto('/audit');
    await page.waitForTimeout(2000);

    // We retry seating traces here inside the test itself in case it was lost in setup
    await request.post('/api/v1/debug/traces', { data: {}, headers: { 'Authorization': 'Bearer test-token', 'Content-Type': 'application/json' } });

    // Wait until loading indicator goes away and 'No logs found' OR 'list-users' appears.
    const noLogs = page.locator('text=No logs found.');
    const listUsers = page.locator('text=list-users').first();

    for (let i = 0; i < 5; i++) {
        if (await listUsers.isVisible()) break;
        if (await noLogs.isVisible()) {
            await page.click('button:has-text("Filter")');
            await page.waitForTimeout(1000);
        }
    }

    await expect(listUsers).toBeVisible({ timeout: 15000 });
    await listUsers.click();

    await expect(page.locator('text=Audit Log Detail')).toBeVisible();
    await expect(page.getByPlaceholder('Search all columns...')).toBeVisible();

    await expect(page.locator('text=Alice').first()).toBeVisible();
    await expect(page.locator('text=admin').first()).toBeVisible();
  });
});
