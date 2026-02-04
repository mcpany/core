/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stacks Management', () => {
    test.afterEach(async ({ request }) => {
        // Cleanup potential leftovers
        await cleanupCollection('New-Stack', request);
        await cleanupCollection('e2e-seed-stack', request);
    });

    test('should list existing stacks', async ({ page, request }) => {
        await seedCollection('e2e-seed-stack', request);
        await page.goto('/stacks');
        await expect(page.getByText('e2e-seed-stack')).toBeVisible();
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
        await expect(page.getByText('e2e-seed-stack')).toBeVisible();

        // Find the card for e2e-seed-stack and click delete
        // My StackCard has Delete button in CardFooter which is hidden until hover.
        // But Playwright can force click or hover.
        const card = page.locator('.group').filter({ hasText: 'e2e-seed-stack' });
        await card.hover();
        await card.getByRole('button', { name: 'Delete' }).click();

        // Confirm dialog
        await expect(page.getByText('Are you absolutely sure?')).toBeVisible();
        await page.getByRole('button', { name: 'Delete Stack' }).click();

        await expect(page.getByText('Stack Deleted', { exact: true })).toBeVisible();
        // Verify Card disappears
        await expect(page.locator('.text-2xl', { hasText: 'e2e-seed-stack' })).toBeHidden();
    });
});
