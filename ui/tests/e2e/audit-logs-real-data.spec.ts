/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Audit Logs Real Data (No Mocks)', () => {
  test('should seed the database and render table for tabular result', async ({ page, request }) => {
    // 1. Seed the trace data in the database
    const res = await request.post('/api/v1/debug/traces');
    expect(res.ok()).toBeTruthy();

    // 2. Go to the audit page
    await page.goto('/audit');

    // 3. Wait for the rows to appear
    await expect(page.getByRole('row').nth(1)).toBeVisible();

    // 4. Click the "View" button for the data-analyzer tool (which has the tabular data)
    // Find the row containing "data-analyzer"
    const row = page.getByRole('row').filter({ hasText: 'search-tool' }).first();
    const viewButton = row.getByRole('button', { name: 'View' });
    await viewButton.click();

    // 5. Verify the Modal opens and shows the expected data in a table
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    await expect(dialog.getByRole('heading', { name: 'Audit Log Detail' })).toBeVisible();

    // The result contains "results": [{"file": ...}] which is a table.
    // RichResultViewer should render a table for it if you click "Table" tab
    const tableTab = dialog.getByRole('tab', { name: 'Table' });
    if (await tableTab.isVisible()) {
      await tableTab.click();
    }
    const table = dialog.getByRole('table').last();
    await expect(table).toBeVisible();
    await expect(table.getByRole('cell', { name: 'report_q3.pdf' })).toBeVisible();
  });
});
