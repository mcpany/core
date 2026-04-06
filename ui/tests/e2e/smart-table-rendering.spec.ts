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

        // Use an even more robust way to click the inspector
        // We will target the "Inspect" button directly, avoiding any reliance on clicking the row text
        // which might be obscured or unclickable.
        // Usually, the Inspect button has text or a title or aria-label.
        // Looking at the table implementation, it's a button with Play icon and "Inspect" text.
        await page.getByRole('button', { name: /inspect/i }).first().click();

        // Wait for inspector - instead of looking for the exact tool name heading,
        // look for the dialog role, which signifies the inspector is open
        await expect(page.getByRole('dialog')).toBeVisible();

        // Ensure we are inside the inspector dialog when we click execute.
        const dialog = page.getByRole('dialog');

        // Execute tool (will be executed against the real backend)
        await dialog.getByRole('button', { name: 'Execute' }).click();

        // Wait for execution to complete
        // Verify that the 'Table' button is available, since the tool returns array of objects
        const tableTab = dialog.getByRole('button', { name: /Table/i });
        await expect(tableTab).toBeVisible({ timeout: 15000 });

        // Click Table view to ensure it's active
        await tableTab.click();

        // Verify the table actually renders
        const table = dialog.locator('table').first();
        await expect(table).toBeVisible();

        // Verify headers (id, name, role, active, long_text)
        await expect(dialog.getByRole('columnheader', { name: 'name' })).toBeVisible();
        await expect(dialog.getByRole('columnheader', { name: 'role' })).toBeVisible();

        // Verify there is at least one row
        const rows = dialog.locator('tbody tr');
        expect(await rows.count()).toBeGreaterThanOrEqual(3);
    });
});
