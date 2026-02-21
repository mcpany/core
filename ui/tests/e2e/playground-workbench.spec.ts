/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices } from './test-data';

test.describe('Playground Workbench', () => {
    test.beforeAll(async () => {
        await seedServices();
    });

    test.afterAll(async () => {
        await cleanupServices();
    });

    test('should open tool in split pane and run it', async ({ page }) => {
        // Navigate to playground
        await page.goto('/playground');

        // Wait for tool list to load and find echo_tool
        // We use a locator that targets the sidebar item specifically to distinguish from potential header
        const toolSidebarItem = page.locator('.group', { hasText: 'echo_tool' });
        await expect(toolSidebarItem).toBeVisible({ timeout: 10000 });
        await toolSidebarItem.click();

        // Verify split pane (Tool Config Panel) exists
        // We look for the specific header structure we added: h2 with text "echo_tool" inside a border-b div
        // And ensure it is NOT inside a dialog
        const configHeader = page.getByRole('heading', { name: 'echo_tool', level: 2 });
        await expect(configHeader).toBeVisible();

        // Assert that no dialog is open (role=dialog is standard for Shadcn Dialog)
        const dialog = page.getByRole('dialog');
        await expect(dialog).not.toBeVisible();

        // Run the tool
        // Since echo_tool has empty schema properties in seed, it should show "This tool takes no arguments"
        // and a "Run Tool" button.
        const runButton = page.getByRole('button', { name: 'Run Tool' });
        await expect(runButton).toBeVisible();
        await runButton.click();

        // Verify result in chat
        // The seed config executes `echo echoed_output`, so we expect "echoed_output" in the result.
        // We look for the result text in the chat area.
        // Note: usage of .first() to handle potential duplicate rendering in JsonView or SmartResultRenderer
        await expect(page.getByText('echoed_output').first()).toBeVisible({ timeout: 10000 });

        // Verify form is still visible (Persistence Check)
        await expect(configHeader).toBeVisible();

        // Close the panel
        const closeButton = page.getByRole('button', { name: 'Close' });
        await expect(closeButton).toBeVisible();
        await closeButton.click();

        // Verify panel is gone
        await expect(configHeader).not.toBeVisible();
    });
});
