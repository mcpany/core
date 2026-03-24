/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Audit Log Viewer', () => {

  test.beforeEach(async ({ request, page }) => {
    // 1. Seed global state (including a test user, services, tools)
    await seedGlobalState(request);

    // 2. Trigger an action that creates an audit log
    // Execute a simple tool (if any is available in the seeded state)
    const _executeRes = await request.post('/api/v1/mcp/execute', {
      data: {
        name: 'test_tool',
        arguments: { arg1: 'value1' }
      }
    });
    // It's ok if the tool execution fails, the audit log should still be recorded.

    // 3. Login
    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    await page.fill('input[name="username"]', 'e2e-admin-core');
    await page.fill('input[name="password"]', 'password');
    await Promise.all([
      page.waitForURL('/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
    ]);
    await expect(page).toHaveURL('/', { timeout: 15000 });
  });

  test('should render audit logs with JsonView tables instead of raw dumps', async ({ page }) => {
    // Navigate to audit page
    await page.goto('/audit');
    await expect(page.getByRole('heading', { name: 'Filters' })).toBeVisible();

    // Wait for the table to populate with at least one log
    // We look for the "View" button which is only rendered for log rows
    const firstViewBtn = page.getByRole('button', { name: 'View' }).first();
    await expect(firstViewBtn).toBeVisible({ timeout: 15000 });

    // Click the View button on the first log
    await firstViewBtn.click();

    // The Dialog should open
    const dialogTitle = page.getByRole('heading', { name: 'Audit Log Detail' });
    await expect(dialogTitle).toBeVisible();

    // Check if JsonView toolbar elements (like Raw or Table buttons) are visible
    // instead of the old SyntaxHighlighter (which had no such buttons).
    // Specifically look for the "Raw" button in the JsonView toolbar.
    const rawBtn = page.getByRole('button', { name: 'Raw' }).first();
    await expect(rawBtn).toBeVisible();
  });

});
