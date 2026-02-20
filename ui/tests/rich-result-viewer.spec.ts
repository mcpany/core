/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('RichResultViewer displays complex data in table view', async ({ page }) => {
  // Go to tools page
  await page.goto('/tools');

  // Filter by our test service/tool
  await page.getByPlaceholder('Search tools...').fill('get_complex_data');

  // Open inspector
  await page.locator('tr').filter({ hasText: 'get_complex_data' }).getByText('Inspect').click();

  // Wait for inspector
  await expect(page.getByText('get_complex_data').first()).toBeVisible();

  // Click Execute
  await page.getByRole('button', { name: 'Execute' }).click();

  // Wait for Result label to appear
  await expect(page.getByText('Result', { exact: true })).toBeVisible();

  // The RichResultViewer should show tabs: "Result" and "Full Output" if parsing succeeded
  await expect(page.getByRole('tab', { name: 'Result' })).toBeVisible();
  await expect(page.getByRole('tab', { name: 'Full Output' })).toBeVisible();

  // By default "Result" tab is selected.
  // Inside it, JsonView should detect table data and show "Table" button/view.

  // Wait for Table button in the toolbar
  await expect(page.getByRole('button', { name: 'Table' })).toBeVisible();

  // Click Table button (it might be auto-selected, but clicking ensures)
  await page.getByRole('button', { name: 'Table' }).click();

  // Check for data in the table
  await expect(page.getByRole('cell', { name: 'Alice' })).toBeVisible();
  await expect(page.getByRole('cell', { name: 'Bob' })).toBeVisible();
  await expect(page.getByRole('cell', { name: 'Admin' })).toBeVisible();
  await expect(page.getByRole('cell', { name: 'User' })).toBeVisible();
});
