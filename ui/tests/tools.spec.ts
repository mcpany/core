/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices } from './e2e/test-data';

test.describe('Tool Exploration', () => {
    // Ensure serial execution to avoid state conflict with other tests
    test.describe.configure({ mode: 'serial' });

    test.beforeAll(async ({ request }) => {
        await seedServices(request);
    });

    test.afterAll(async ({ request }) => {
        await cleanupServices(request);
    });

    test('should list available tools from real backend', async ({ page }) => {
        await page.goto('/tools');
        // Check for tools seeded by seedServices
        await expect(page.getByText('process_payment').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Process a payment').first()).toBeVisible({ timeout: 10000 });

        await expect(page.getByText('calculator').first()).toBeVisible({ timeout: 10000 });
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');
        const toolRow = page.locator('tr').filter({ hasText: 'process_payment' });
        await toolRow.getByText('Inspect').click();

        await expect(page.getByText('Schema', { exact: true }).first()).toBeVisible();
        await expect(page.getByText('process_payment').first()).toBeVisible();
    });
});
