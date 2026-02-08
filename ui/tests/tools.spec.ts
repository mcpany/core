/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser, cleanupUser } from './e2e/test-data';
import { login } from './e2e/auth-helper';

test.describe('Tool Exploration', () => {
    test.beforeEach(async ({ page, request }) => {
        await seedUser(request, "e2e-admin");
        await seedServices(request);
        await login(page);
    });

    test.afterEach(async ({ request }) => {
        await cleanupServices(request);
        await cleanupUser(request, "e2e-admin");
    });

    test('should list available tools', async ({ page }) => {
        await page.goto('/tools');
        await expect(page.getByText('calculator').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('process_payment').first()).toBeVisible({ timeout: 10000 });
    });

    test('should show empty state when no tools', async ({ page, request }) => {
        // We need to cleanup services to simulate empty state
        await cleanupServices(request);
        await page.reload();
        await expect(page.locator('table tbody tr')).toHaveCount(1, { timeout: 15000 });
        await expect(page.locator('text=No tools found.')).toBeVisible();
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');
        const toolRow = page.locator('tr').filter({ hasText: 'calculator' });
        await toolRow.getByText('Inspect').click({ timeout: 15000 });

        await expect(page.getByText('Schema', { exact: true }).first()).toBeVisible({ timeout: 15000 });
        await expect(page.getByText('calculator').first()).toBeVisible();
    });
});
