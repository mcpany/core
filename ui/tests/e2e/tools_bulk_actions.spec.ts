/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Tools Bulk Actions', () => {
    test.beforeAll(async ({ request }) => {
        // Seed a service with tools using the real API
        const response = await request.post('/api/v1/services', {
            data: {
                id: "bulk-test-service",
                name: "bulk-test-service",
                version: "1.0.0",
                priority: 0,
                disable: false,
                command_line_service: {
                    command: "echo",
                    working_directory: "/tmp",
                    tools: [
                        {
                            name: "bulk-tool-1",
                            description: "Test tool 1 for bulk actions",
                            inputSchema: { type: "object", properties: { arg: { type: "string" } } }
                        },
                        {
                            name: "bulk-tool-2",
                            description: "Test tool 2 for bulk actions",
                            inputSchema: { type: "object", properties: { arg: { type: "string" } } }
                        },
                        {
                            name: "bulk-tool-3",
                            description: "Test tool 3 for bulk actions",
                            inputSchema: { type: "object", properties: { arg: { type: "string" } } }
                        }
                    ]
                }
            }
        });

        // Ensure creation was successful (or already exists)
        expect(response.ok() || response.status() === 409).toBeTruthy();
    });

    test('should allow bulk disabling of tools', async ({ page }) => {
        // Go to Tools page
        await page.goto('/tools');

        // Wait for tools to load
        await expect(page.getByText('bulk-tool-1')).toBeVisible();
        await expect(page.getByText('bulk-tool-2')).toBeVisible();
        await expect(page.getByText('bulk-tool-3')).toBeVisible();

        // Initially they should be enabled
        await expect(page.locator('tr', { hasText: 'bulk-tool-1' }).getByText('Enabled')).toBeVisible();

        // Select tool 1 and 2 (using the new checkboxes we will add)
        // We assume the checkbox will have an aria-label or accessible role
        // For now, let's target the checkbox in the first cell
        await page.locator('tr', { hasText: 'bulk-tool-1' }).getByRole('checkbox').check();
        await page.locator('tr', { hasText: 'bulk-tool-2' }).getByRole('checkbox').check();

        // Verify Bulk Action Bar appears
        await expect(page.getByText('2 selected')).toBeVisible();

        // Click "Disable"
        await page.getByRole('button', { name: 'Disable' }).click();

        // Verify UI updates
        // tool-1 and tool-2 should now be Disabled
        await expect(page.locator('tr', { hasText: 'bulk-tool-1' }).getByText('Disabled')).toBeVisible();
        await expect(page.locator('tr', { hasText: 'bulk-tool-2' }).getByText('Disabled')).toBeVisible();

        // tool-3 should still be Enabled
        await expect(page.locator('tr', { hasText: 'bulk-tool-3' }).getByText('Enabled')).toBeVisible();

        // Verify backend state via API
        const response = await page.request.get('/api/v1/tools');
        const data = await response.json();
        const tools = data.tools || [];

        const tool1 = tools.find((t: any) => t.name === 'bulk-tool-1');
        const tool2 = tools.find((t: any) => t.name === 'bulk-tool-2');
        const tool3 = tools.find((t: any) => t.name === 'bulk-tool-3');

        expect(tool1.disable).toBe(true);
        expect(tool2.disable).toBe(true);
        expect(tool3.disable).toBeFalsy();
    });

    test('should allow bulk enabling of tools', async ({ page }) => {
        await page.goto('/tools');

        // Ensure state from previous test is handled or reset (but we are running sequential)
        // Let's assume we start from the disabled state

        // Select tool 1
        await page.locator('tr', { hasText: 'bulk-tool-1' }).getByRole('checkbox').check();

        // Click "Enable"
        await page.getByRole('button', { name: 'Enable' }).click();

        // Verify UI updates
        await expect(page.locator('tr', { hasText: 'bulk-tool-1' }).getByText('Enabled')).toBeVisible();
    });

    test('should select all tools', async ({ page }) => {
        await page.goto('/tools');
        await expect(page.getByText('bulk-tool-1')).toBeVisible();

        // Click "Select All" in header
        await page.locator('thead').getByRole('checkbox').check();

        // Verify all rows are checked
        await expect(page.locator('tr', { hasText: 'bulk-tool-1' }).getByRole('checkbox')).toBeChecked();
        await expect(page.locator('tr', { hasText: 'bulk-tool-2' }).getByRole('checkbox')).toBeChecked();
        await expect(page.locator('tr', { hasText: 'bulk-tool-3' }).getByRole('checkbox')).toBeChecked();

        // Verify count in action bar
        // Note: The count might include other tools if the environment is dirty.
        // We should check that it is at least 3.
        const countText = await page.getByText(/selected/).textContent();
        const count = parseInt(countText?.split(' ')[0] || '0');
        expect(count).toBeGreaterThanOrEqual(3);
    });
});
