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
    await page.goto('/audit');
    await page.waitForTimeout(1000);

    // We retry seating traces here inside the test itself in case it was lost in setup
    try {
      await request.post('/api/v1/debug/traces', { data: {}, headers: { 'Authorization': 'Bearer test-token', 'Content-Type': 'application/json' } });
    } catch(e) {}

    const listUsers = page.locator('text=list-users').first();
    const noLogs = page.locator('text=No logs found.');

    // A more active loop checking for logs in case websocket fails and requires explicit filtering
    for (let i = 0; i < 5; i++) {
        if (await listUsers.isVisible()) break;
        try {
            await page.click('button:has-text("Filter")');
        } catch(e) {}
        await page.waitForTimeout(1500);
    }

    await expect(listUsers).toBeVisible({ timeout: 15000 });
    await listUsers.click();

    await expect(page.locator('text=Audit Log Detail')).toBeVisible();
    await expect(page.getByPlaceholder('Search all columns...')).toBeVisible();

    await expect(page.locator('text=Alice').first()).toBeVisible();
    await expect(page.locator('text=admin').first()).toBeVisible();
  });
});
