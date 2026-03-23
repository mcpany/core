/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Inspector Detailed View (Rich Payload)', () => {
  test.beforeEach(async ({ page, request }) => {
    await seedGlobalState(request);

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

  test('should display RichResultViewer for payloads in Inspector traces', async ({ page, request }) => {
    // Seed real trace with tabular output data
    const tabularOutput = [
      { id: 1, name: "Alice", active: true },
      { id: 2, name: "Bob", active: false }
    ];

    const testTrace = {
        id: "e2e-trace-rich-payload-test",
        rootSpan: {
            id: "span-rich",
            name: "fetch_users",
            serviceName: "User DB",
            type: "tool",
            status: "success",
            startTime: Date.now() - 200,
            endTime: Date.now(),
            children: [],
            input: { query: "SELECT * FROM users;" },
            output: tabularOutput,
        },
        timestamp: new Date().toISOString(),
        totalDuration: 200,
        status: "success",
        trigger: "user"
    };

    // Use actual backend to seed trace
    const response = await request.post('/api/v1/debug/traces', {
      data: testTrace,
      headers: {
        'X-API-Key': 'test-token',
        'Content-Type': 'application/json'
      }
    });
    expect(response.ok()).toBeTruthy();

    await page.goto('/inspector');

    // Wait for Inspector Table to load our specific seeded trace
    const ourTraceRow = page.locator('table').locator('tr').filter({ hasText: 'e2e-trace-rich-payload-test' }).first();
    await expect(ourTraceRow).toBeVisible({ timeout: 15000 });

    // Click our trace to open details pane
    await ourTraceRow.click();

    // Check if details pane is open by checking for trace name
    await expect(page.locator('.text-2xl.font-bold', { hasText: 'fetch_users' })).toBeVisible();

    // Click the Payload tab
    const payloadTab = page.getByRole('tab', { name: 'Payload' });
    await expect(payloadTab).toBeVisible();
    await payloadTab.click();

    // Verify RichResultViewer rendered the table structure
    await expect(page.locator('h3', { hasText: 'Response Payload' })).toBeVisible();

    // Check if "Table" tab in RichResultViewer is selected or visible
    // Depending on RichResultViewer implementation, we look for table elements
    const tableHeader = page.locator('th', { hasText: 'name' });
    await expect(tableHeader).toBeVisible();

    const aliceCell = page.locator('td', { hasText: 'Alice' });
    await expect(aliceCell).toBeVisible();
    const bobCell = page.locator('td', { hasText: 'Bob' });
    await expect(bobCell).toBeVisible();
  });
});
