/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stacks List', () => {
    const stackName = 'e2e-stack-test';

    test.beforeEach(async ({ request }) => {
        // Ensure clean state
        await cleanupCollection(stackName, request);
        // Seed
        await seedCollection(stackName, request);
    });

    test.afterEach(async ({ request }) => {
        await cleanupCollection(stackName, request);
    });

    test('should list created stacks', async ({ page }) => {
        await page.goto('/stacks');

        // Check if the stack card is visible
        // The card title (h3 or similar) might be "Stack" but the content has the name
        const stackCard = page.locator('.group').filter({ hasText: stackName });
        await expect(stackCard).toBeVisible({ timeout: 10000 });

        // Click on it to ensure link works
        await stackCard.click();
        await expect(page).toHaveURL(new RegExp(`/stacks/${stackName}`));
    });
});
