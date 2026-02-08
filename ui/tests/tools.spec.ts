/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices } from './e2e/test-data';

test.describe('Tool Exploration', () => {
    test.beforeEach(async ({ request }) => {
        await seedServices(request);
    });

    test.afterEach(async ({ request }) => {
        await cleanupServices(request);
    });

    test('should list available tools', async ({ page }) => {
        await page.goto('/tools');

        // Expect seeded tools (Memory) - Increase timeout for npx installation
        await expect(page.getByText('read_graph').first()).toBeVisible({ timeout: 45000 });
    });

    test('should show empty state when no tools', async ({ page, request }) => {
        // Cleanup tools for this test
        await cleanupServices(request);
        // Cleanup Memory service manually if cleanupServices didn't catch it (redundant if test-data.ts updated, but safe)
        await request.delete('/api/v1/services/Memory', { headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' } }).catch(() => {});

        await page.goto('/tools');
        await page.reload(); // Ensure fresh state

        // The table shows one row with "No tools found." when empty
        await expect(page.locator('text=No tools found.')).toBeVisible({ timeout: 15000 });
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');
        // Wait for tool to appear first
        await expect(page.getByText('read_graph').first()).toBeVisible({ timeout: 45000 });

        const toolRow = page.locator('tr').filter({ hasText: 'read_graph' });
        await toolRow.getByText('Inspect').click();

        await expect(page.getByText('Schema', { exact: true }).first()).toBeVisible();
        await expect(page.getByText('read_graph').first()).toBeVisible();
    });
});
