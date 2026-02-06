/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stacks List', () => {
    test.beforeEach(async ({ request }) => {
        // Cleanup potential leftovers
        await cleanupCollection('stack-alpha', request);
        await cleanupCollection('stack-beta', request);
        await cleanupCollection('stack-gamma', request);

        await seedCollection('stack-alpha', request);
        await seedCollection('stack-beta', request);
    });

    test.afterEach(async ({ request }) => {
        await cleanupCollection('stack-alpha', request);
        await cleanupCollection('stack-beta', request);
        await cleanupCollection('stack-gamma', request);
    });

    test('should list seeded stacks', async ({ page }) => {
        await page.goto('/stacks');

        // Check for the stack cards
        await expect(page.getByText('stack-alpha', { exact: true })).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('stack-beta', { exact: true })).toBeVisible();
    });

    test('should create a new stack', async ({ page }) => {
        await page.goto('/stacks');

        await page.getByRole('button', { name: 'Create Stack' }).click();

        // Fill in the name
        await page.getByPlaceholder('e.g. my-stack').fill('stack-gamma');
        await page.getByRole('button', { name: 'Create', exact: true }).click();

        // Expect redirection to the editor or at least to the stack detail
        // The URL should contain /stacks/stack-gamma
        await expect(page).toHaveURL(/\/stacks\/stack-gamma/);
    });

    test('should delete a stack', async ({ page }) => {
        await page.goto('/stacks');

        // Ensure stack-alpha is there
        await expect(page.getByText('stack-alpha', { exact: true })).toBeVisible();

        // Handle confirmation dialog
        page.on('dialog', dialog => dialog.accept());

        // Click delete on stack-alpha
        // We need to find the card that contains 'stack-alpha' and then find the delete button within it.
        // Assuming the card has some unique text 'stack-alpha'
        const card = page.locator('.stack-card').filter({ hasText: 'stack-alpha' });

        // Wait for card to be visible
        await expect(card).toBeVisible();

        // Hover the card to reveal the delete button
        await card.hover();

        // Click the delete button
        await card.getByTitle('Delete Stack').click();

        // Wait for it to disappear
        await expect(page.getByText('stack-alpha', { exact: true })).not.toBeVisible();
    });
});
