/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Tool Exploration', () => {
    test.beforeEach(async ({ page }) => {
        // No mocks - using real backend
    });

    test('should list available tools', async ({ page }) => {
        await page.goto('/tools');

        // Check for 'get_weather' tool which should be available from weather-service
        const toolName = 'get_weather';
        await expect(page.getByText(toolName).first()).toBeVisible({ timeout: 10000 });

        // We can't strictly assert description unless we know it, but logic implies it exists
    });

    test.skip('should show empty state when no tools', async ({ page }) => {
        // Skipped: Cannot ensure empty backend state in shared E2E environment
        await page.goto('/tools');
        await expect(page.locator('table tbody tr')).toHaveCount(0);
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');

        const toolName = 'get_weather';
        const toolRow = page.locator('tr').filter({ hasText: toolName });
        await toolRow.getByText('Inspect').click();

        await expect(page.getByText('Schema', { exact: true }).first()).toBeVisible();
        await expect(page.getByText(toolName).first()).toBeVisible();
    });
});
