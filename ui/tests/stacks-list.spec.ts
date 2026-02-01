/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stacks List', () => {
    const stackName = 'e2e-test-stack-list';

    test.beforeEach(async ({ request }) => {
        // Ensure clean state
        await cleanupCollection(stackName, request);
    });

    test.afterEach(async ({ request }) => {
        await cleanupCollection(stackName, request);
        await cleanupCollection('e2e-new-stack', request);
    });

    test('should display seeded stack', async ({ page, request }) => {
        await seedCollection(stackName, request);

        await page.goto('/stacks');

        // Wait for loading to finish (Loader2 icon should disappear)
        await expect(page.locator('.lucide-loader-2')).not.toBeVisible();

        // Check if stack card exists with correct name
        await expect(page.getByText(stackName, { exact: true })).toBeVisible();
    });

    test('should create a new stack', async ({ page }) => {
        const newStackName = 'e2e-new-stack';
        // Ensure cleanup first

        await page.goto('/stacks');

        // Click Create Stack
        await page.getByRole('button', { name: 'Create Stack' }).click();

        // Fill dialog
        await page.getByLabel('Name').fill(newStackName);
        await page.getByRole('button', { name: 'Create', exact: true }).click();

        // Verify redirection to details page
        await expect(page).toHaveURL(new RegExp(`/stacks/${newStackName}`));

        // Verify we are on the details page (check for title)
        await expect(page.getByRole('heading', { name: newStackName })).toBeVisible();
    });
});
