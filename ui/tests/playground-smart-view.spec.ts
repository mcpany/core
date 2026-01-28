/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Smart Result Visualization', () => {
    // We rely on 'user-service' seeded via config.minimal.yaml
    const toolName = 'user-service.list_users';

    test('should render JSON array as a table and allow toggling to raw view', async ({ page }) => {
        await page.goto('/playground');

        // Wait for the tool to be available (poll or search)
        // We can type the tool name directly into the input
        const input = page.getByPlaceholder('Enter command or select a tool...');
        await input.fill(`${toolName} {}`);
        await input.press('Enter');

        // Wait for result
        // The result should contain a table
        // We look for the "Table" button which indicates Smart View is active
        await expect(page.getByRole('button', { name: 'Table' })).toBeVisible({ timeout: 15000 });
        await expect(page.getByRole('button', { name: 'JSON' })).toBeVisible();

        // Verify Table headers
        await expect(page.getByRole('columnheader', { name: 'id' })).toBeVisible();
        await expect(page.getByRole('columnheader', { name: 'name' })).toBeVisible();
        await expect(page.getByRole('columnheader', { name: 'role' })).toBeVisible();

        // Verify Table content
        await expect(page.getByRole('cell', { name: 'Alice' })).toBeVisible();
        await expect(page.getByRole('cell', { name: 'Bob' })).toBeVisible();

        // Toggle to Raw JSON
        await page.getByRole('button', { name: 'JSON' }).click();

        // Verify Table is gone
        await expect(page.getByRole('columnheader', { name: 'id' })).not.toBeVisible();

        // Verify JSON content is visible
        await expect(page.locator('.group\\/code').last()).toBeVisible();
        await expect(page.locator('.group\\/code').last()).toContainText('Alice');

        // Toggle back to Table
        await page.getByRole('button', { name: 'Table' }).click();
        await expect(page.getByRole('columnheader', { name: 'name' })).toBeVisible();
    });
});
