/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Audit Logs Real Data (No Mocks)', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
        localStorage.setItem('mcp_auth_token', 'test-token');
    });
  });

  test('should seed the database and render table for tabular result', async ({ page, request }) => {
    // 1. Seed the trace data in the database using the API with the test token
    const res = await request.post('/api/v1/debug/traces', {
      headers: {
        'Authorization': 'Bearer test-token'
      }
    });

    if (!res.ok()) {
      const errorText = await res.text();
      console.error(`Failed to seed trace: ${res.status()} ${errorText}`);
    }
    expect(res.ok()).toBeTruthy();

    // 2. Go to the audit page
    await page.goto('/audit');

    // 3. Wait for the rows to appear
    await expect(page.getByText('search-tool').first()).toBeVisible({ timeout: 10000 });

    // 4. Click the "View" button for the search-tool
    const row = page.locator('tr').filter({ hasText: 'search-tool' }).first();
    await row.getByRole('button', { name: 'View' }).click();

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
