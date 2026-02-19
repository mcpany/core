/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('Rich Result Viewer displays complex data as table', async ({ page }) => {
  // Ensure we are using the real backend by NOT mocking /api/v1/tools

  await page.goto('/tools');

  // Search for the tool
  await page.getByPlaceholder('Search tools...').fill('get_complex_data');

  // Wait for the tool to appear
  await expect(page.getByText('get_complex_data')).toBeVisible();

  // Open inspector
  // Depending on the view (card or table), the button might be different.
  // Assuming default view or consistent selector.
  // In `tool-inspector.spec.ts`: await page.locator('tr').filter({ hasText: 'get_weather' }).getByText('Inspect').click();
  // But if it's card view? The ToolTable component handles both or just table?
  // Let's assume table view is default or we can switch.
  // The `ToolTable` component renders table rows.
  await page.locator('tr').filter({ hasText: 'get_complex_data' }).getByRole('button', { name: 'Inspect' }).click();

  // Wait for inspector to open
  await expect(page.getByRole('dialog')).toBeVisible();
  await expect(page.getByRole('heading', { name: 'get_complex_data' })).toBeVisible();

  // Click Execute (default args are empty object which is fine for this tool if schema allows,
  // or we might need to verify schema form is shown)
  // The tool definition in config.test.yaml has empty properties, so {} is valid.
  await page.getByRole('button', { name: 'Execute' }).click();

  // Wait for result
  // The result label should be visible
  await expect(page.getByText('Result')).toBeVisible();

  // Check if Table view is active (JsonView defaults to smart table if array of objects)
  // The Table button should be active (secondary variant)
  await expect(page.getByRole('button', { name: 'Table' })).toHaveClass(/bg-secondary/);

  // Check table content
  // We expect "Alice" and "Bob" to be visible in the table cells
  await expect(page.getByRole('cell', { name: 'Alice' })).toBeVisible();
  await expect(page.getByRole('cell', { name: 'Bob' })).toBeVisible();
  await expect(page.getByRole('cell', { name: 'Admin' })).toBeVisible();

  // Switch to Raw view
  await page.getByRole('button', { name: 'Raw' }).click();

  // Check raw JSON content
  await expect(page.locator('code').filter({ hasText: '"name": "Alice"' })).toBeVisible();
});
