/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stacks Management', () => {

  test('should list existing stacks', async ({ page, request }) => {
    const stackName = 'e2e-test-stack-list';
    await seedCollection(stackName, request);

    try {
        await page.goto('/stacks');
        // Wait for loading to finish
        await expect(page.locator('.animate-spin')).toHaveCount(0);
        await expect(page.getByText(stackName)).toBeVisible();
    } finally {
        await cleanupCollection(stackName, request);
    }
  });

  test('should create a new stack', async ({ page, request }) => {
    const stackName = 'e2e-new-stack';
    // Ensure cleanup in case of previous failure
    await cleanupCollection(stackName, request);

    try {
        await page.goto('/stacks');
        // Wait for loading to finish
        await expect(page.locator('.animate-spin')).toHaveCount(0);

        await page.getByText('Add Stack').click();
        await page.getByLabel('Name').fill(stackName);
        await page.getByRole('button', { name: 'Create' }).click();

        // Should verify success
        await expect(page.locator('div.group', { hasText: stackName })).toBeVisible();

    } finally {
        await cleanupCollection(stackName, request);
    }
  });

  test('should delete a stack', async ({ page, request }) => {
    const stackName = 'e2e-test-stack-delete';
    await seedCollection(stackName, request);

    try {
        await page.goto('/stacks');
        // Wait for loading to finish
        await expect(page.locator('.animate-spin')).toHaveCount(0);
        await expect(page.getByText(stackName)).toBeVisible();

        const card = page.locator('.group', { hasText: stackName });
        await expect(card).toBeVisible();

        // Click dropdown trigger (MoreHorizontal icon button)
        await card.getByRole('button').click();

        // Click delete in menu
        await page.getByRole('menuitem', { name: 'Delete' }).click();

        // Verify AlertDialog is visible
        await expect(page.getByRole('alertdialog')).toBeVisible();

        // Click Delete in AlertDialog
        await page.getByRole('button', { name: 'Delete' }).click();

        // Verify it disappears
        await expect(card).toHaveCount(0);
    } finally {
        await cleanupCollection(stackName, request);
    }
  });
});
