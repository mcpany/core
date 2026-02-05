/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection, seedUser, cleanupUser } from './e2e/test-data';

test.describe('Stacks Management', () => {
    test.beforeEach(async ({ page, request }) => {
        // Seed admin user for auth
        await seedUser(request, "admin");
        // Set auth token (basic auth for admin:password)
        await page.addInitScript(() => {
            localStorage.setItem('mcp_auth_token', btoa('admin:password'));
        });
    });

    test.afterEach(async ({ request }) => {
        // Cleanup potential leftovers
        await cleanupCollection('New-Stack', request);
        await cleanupCollection('e2e-seed-stack', request);
        await cleanupUser(request, "admin");
    });

    test('should list existing stacks', async ({ page, request }) => {
        await seedCollection('e2e-seed-stack', request);
        await page.goto('/stacks');
        await expect(page.locator('.text-2xl', { hasText: 'e2e-seed-stack' })).toBeVisible();
    });

    test('should create a new stack', async ({ page }) => {
        await page.goto('/stacks');
        await page.getByRole('button', { name: 'Create Stack' }).click();

        await page.getByLabel('Name').fill('New-Stack');
        await page.getByLabel('Description').fill('Created by E2E');
        await page.getByRole('button', { name: 'Create', exact: true }).click();

        await expect(page.getByText('Stack Created', { exact: true })).toBeVisible();
        // Verify Card appears
        await expect(page.locator('.text-2xl', { hasText: 'New-Stack' })).toBeVisible();
    });

    test('should delete a stack', async ({ page, request }) => {
        await seedCollection('e2e-seed-stack', request);
        await page.goto('/stacks');

        // Wait for list to load
        const card = page.locator('.group').filter({ hasText: 'e2e-seed-stack' });
        await expect(card).toBeVisible({ timeout: 10000 });

        // Force hover to show actions
        await card.hover();
        // Force click if visibility is tricky (CSS opacity)
        await card.getByRole('button', { name: 'Delete' }).click({ force: true });

        // Confirm dialog
        await expect(page.getByText('Are you absolutely sure?')).toBeVisible();
        await page.getByRole('button', { name: 'Delete Stack' }).click();

        await expect(page.getByText('Stack Deleted', { exact: true })).toBeVisible();
        // Verify Card disappears
        await expect(page.locator('.text-2xl', { hasText: 'e2e-seed-stack' })).toBeHidden();
    });
});
