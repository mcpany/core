/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stacks List', () => {
    test.beforeEach(async ({ request }) => {
        // Seed a stack to verify it appears in the list
        await seedCollection('test-stack-e2e', request);
    });

    test.afterEach(async ({ request }) => {
        // Cleanup
        await cleanupCollection('test-stack-e2e', request);
        await cleanupCollection('new-stack-ui', request);
    });

    test('should list seeded stacks', async ({ page }) => {
        await page.goto('/stacks');

        // Check if the seeded stack card exists
        await expect(page.getByText('test-stack-e2e', { exact: true })).toBeVisible({ timeout: 10000 });

        // Click on it and verify navigation
        await page.getByText('test-stack-e2e', { exact: true }).click();
        await expect(page).toHaveURL(/\/stacks\/test-stack-e2e/);
    });

    test('should create a new stack', async ({ page }) => {
        await page.goto('/stacks');

        // Click create button
        await page.getByRole('button', { name: 'Create Stack' }).click();

        // Fill form
        const dialog = page.getByRole('dialog');
        await expect(dialog).toBeVisible();
        await dialog.getByLabel('Name').fill('new-stack-ui');
        await dialog.getByLabel('Description').fill('Created via E2E test');

        // Submit
        await dialog.getByRole('button', { name: 'Create Stack' }).click();

        // Expect toast or dialog close
        await expect(dialog).toBeHidden();

        // Verify it appears in list
        await expect(page.getByText('new-stack-ui', { exact: true })).toBeVisible();
    });
});
