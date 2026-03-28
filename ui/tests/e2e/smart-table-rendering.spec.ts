/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('SmartTable Rendering and Real Data Validation', () => {
    test.beforeEach(async ({ request }) => {
        await seedGlobalState(request);
    });

    test('should render SmartTable for tool returning array of objects without mocking', async ({ page }) => {
        // We will execute a tool that returns tabular data seeded in the database
        await page.goto('/tools');

        // Ensure Tools Page loaded
        await expect(page.getByRole('heading', { name: 'Tools' })).toBeVisible({ timeout: 15000 });

        // Go to a known tool, 'get_users', seeded in test-data.ts
        const toolName = 'get_users';

        // Find the tool in the list
        await page.getByPlaceholder('Search tools...').fill(toolName);

        // Let's use a very permissive locator since the table cell might contain icons/badges
        // We'll look for a table cell that contains the toolName text and click it
        const toolRow = page.locator('tr', { hasText: toolName }).first();
        await expect(toolRow).toBeVisible({ timeout: 15000 });

        // Click the button inside the row that opens the inspector, or just the row if it's clickable
        // Usually there's an "Inspect" or "Play" button in the actions column. Let's click that to be safe.
        await toolRow.getByRole('button', { name: /inspect/i }).click();

        // Wait for inspector
        await expect(page.getByRole('dialog')).toBeVisible();
        await expect(page.getByRole('heading', { name: toolName, exact: true })).toBeVisible();

        // Execute tool (will be executed against the real backend)
        await page.getByRole('button', { name: 'Execute' }).click();

        // Wait for execution to complete
        // Verify that the 'Table' tab is active or available, since the tool returns array of objects
        const tableTab = page.getByRole('tab', { name: 'Table' });
        await expect(tableTab).toBeVisible({ timeout: 15000 });

        // Check if Table is selected by default, if not click it
        if (await tableTab.getAttribute('aria-selected') !== 'true') {
            await tableTab.click();
        }

        // Verify the table actually renders
        const table = page.locator('table').first();
        await expect(table).toBeVisible();

        // Verify headers (id, name, role, active, long_text)
        await expect(page.getByRole('columnheader', { name: 'name' })).toBeVisible();
        await expect(page.getByRole('columnheader', { name: 'role' })).toBeVisible();

        // Verify there is at least one row
        const rows = page.locator('tbody tr');
        expect(await rows.count()).toBeGreaterThanOrEqual(3);
    });
});
