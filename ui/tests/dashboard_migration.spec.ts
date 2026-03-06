/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('dashboard layout schema migration', async ({ page, request }) => {
  // Clear API preferences to ensure fallback to localStorage
  await request.post('/api/v1/user/preferences', {
      data: { "dashboard-layout": "[]" }
  });

  // Inject legacy schema into localStorage before page load
  await page.addInitScript(() => {
    const legacyLayout = [
        { id: "metrics", title: "Metrics Overview", type: "wide" }
    ];
    window.localStorage.setItem("dashboard-layout", JSON.stringify(legacyLayout));
  });

  await page.goto('/');

  // Wait for the grid to render the migrated widget
  await expect(page.locator('.animate-spin')).not.toBeVisible();
  await expect(page.getByText('Metrics Overview')).toBeVisible();

  // The code immediately saves the migrated layout to the API
  // Let's verify the API now has the newly migrated format (which must contain instanceId)

  // Wait a short moment for the fetch to resolve
  await page.waitForTimeout(500);

  const response = await request.get('/api/v1/user/preferences');
  expect(response.ok()).toBeTruthy();
  const data = await response.json();

  const savedLayoutStr = data['dashboard-layout'];
  expect(savedLayoutStr).toBeDefined();

  const savedLayout = JSON.parse(savedLayoutStr);

  // It should be an array with 1 item
  expect(Array.isArray(savedLayout)).toBe(true);
  expect(savedLayout.length).toBe(1);

  // It should be migrated to the new schema
  expect(savedLayout[0].instanceId).toBeDefined();
  expect(savedLayout[0].type).toBe('metrics');
  expect(savedLayout[0].size).toBe('full'); // 'wide' maps to 'full'
  expect(savedLayout[0].hidden).toBe(false);
});
