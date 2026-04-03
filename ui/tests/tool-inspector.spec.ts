/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './e2e/test-data';

test.describe('Tool Inspector', () => {
    test.beforeEach(async ({ request }) => {
        await seedGlobalState(request);
    });

    test('Tools page loads and inspector opens with real data', async ({ page }) => {
        await page.goto('/tools');
        await expect(page.getByRole('heading', { name: 'Tools' })).toBeVisible({ timeout: 15000 });

        const toolName = 'echo_tool';
        await page.getByPlaceholder('Search tools...').fill(toolName);

        const toolRow = page.locator('tr').filter({ hasText: toolName });
        await expect(toolRow).toBeVisible({ timeout: 15000 });

        await toolRow.getByRole('button', { name: /inspect/i }).click();
        const dialog = page.getByRole('dialog');
        await expect(dialog).toBeVisible();

        // Verify tool details - use h2 to be specific
        await expect(dialog.locator('h2').filter({ hasText: toolName }).first()).toBeVisible();

        // Navigate to Schema tab
        await dialog.getByRole('tab', { name: 'Schema' }).click();

        // In the Schema tab, there are sub-tabs: Visual and JSON
        // Based on ToolRunner.tsx, it's a simple Tabs component inside Schema tab
        await dialog.getByRole('tab', { name: 'JSON' }).click();

        // Verify the pre tag with JSON content appears
        await expect(dialog.locator('pre').first()).toBeVisible();
    });
});
