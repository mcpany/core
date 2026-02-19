/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('Rich Result Viewer displays complex data as table', async ({ page }) => {
  // 1. Navigate to /tools
  await page.goto('/tools');

  // 2. Find get_complex_data
  // Wait for tools to load (real data from backend). Use regex to match namespaced tools (e.g. test-service.get_complex_data)
  await expect(page.getByText(/get_complex_data/)).toBeVisible({ timeout: 10000 });

  // 3. Open Inspector
  // Find the row with get_complex_data and click Inspect button
  await page.locator('tr').filter({ hasText: /get_complex_data/ }).getByRole('button', { name: 'Inspect' }).click();

  // 4. Wait for inspector to open
  await expect(page.getByRole('dialog')).toBeVisible();
  await expect(page.getByRole('dialog').getByText(/get_complex_data/).first()).toBeVisible();

  // 5. Execute
  // Click "Execute" button.
  await page.getByRole('button', { name: 'Execute' }).click();

  // 6. Check for Table view
  // Wait for result to appear. JsonView renders "Table" button if smart table is detected.
  // The backend echoes: [{"id": 1, "name": "Alice", ...}, ...]
  // So it should be detected as smart table.

  // Wait for "Table" button to be visible in the JsonView toolbar
  await expect(page.getByRole('button', { name: 'Table' })).toBeVisible({ timeout: 10000 });

  // Click Table view if not already selected (JsonView defaults to Smart if detected)
  // It should be selected by default, so we just check for content.

  // Check if table content is visible
  await expect(page.getByRole('cell', { name: 'Alice' })).toBeVisible();
  await expect(page.getByRole('cell', { name: 'Bob' })).toBeVisible();
  await expect(page.getByRole('cell', { name: 'Admin' })).toBeVisible();
});
