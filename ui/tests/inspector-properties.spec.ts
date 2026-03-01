/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Inspector PropertiesTable', () => {
  test('seed trace and view formatted payload', async ({ page, request }) => {
    // 1. Seed a trace using the debug endpoint
    const response = await request.post('/api/v1/debug/traces', {
      data: {}
    });
    expect(response.ok()).toBeTruthy();

    // 2. Navigate to the Inspector UI
    await page.goto('/inspector');

    // Wait for websocket/loading to finish and trace to appear
    await page.waitForSelector('text=orchestrator-task', { timeout: 10000 });

    // 3. Click on the seeded trace row
    await page.locator('text=orchestrator-task').first().click();

    // 4. Open the "Payload" tab in the Sheet
    await page.getByRole('tab', { name: 'Payload' }).click();

    // 5. Verify the new table format renders (PropertiesTable)
    // Check that headers exist
    await page.waitForSelector('text=Property');
    await page.waitForSelector('text=Value');

    // Check specific seeded values (e.g., from generateMockTrace)
    await expect(page.locator('span', { hasText: 'query' }).first()).toBeVisible();
    await expect(page.locator('span', { hasText: 'summary' }).first()).toBeVisible();
    await expect(page.locator('span', { hasText: '"Revenue up 15%"' }).first()).toBeVisible();
  });
});
