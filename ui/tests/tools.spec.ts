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

        // Expect seeded tools (Memory)
        await expect(page.getByText('read_graph').first()).toBeVisible({ timeout: 10000 });
    });

    test('should show empty state when no tools', async ({ page, request }) => {
        // Cleanup tools for this test
        await cleanupServices(request);
        // Also ensure no other tools are lingering (Memory service ID is not in cleanupServices helper yet? Check test-data.ts)
        // seedServices adds "Memory", cleanupServices removes "Payment Gateway", "User Service", "Math".
        // I need to update cleanupServices in test-data.ts or handle it here.
        // Assuming I updated test-data.ts, but I didn't update cleanupServices to include Memory!
        // I must update test-data.ts first or do manual cleanup.
        // Let's do manual cleanup for Memory here to be safe or update test-data.ts in next step?
        // Wait, I can update test-data.ts now? No, I am editing tools.spec.ts.
        // I will use raw request to delete Memory service.
        await request.delete('/api/v1/services/Memory', { headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' } });

        await page.goto('/tools');
        await page.reload(); // Ensure fresh state

        // The table shows one row with "No tools found." when empty
        await expect(page.locator('text=No tools found.')).toBeVisible({ timeout: 10000 });
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');
        const toolRow = page.locator('tr').filter({ hasText: 'read_graph' });
        await toolRow.getByText('Inspect').click();

        await expect(page.getByText('Schema', { exact: true }).first()).toBeVisible();
        await expect(page.getByText('read_graph').first()).toBeVisible();
    });
});
