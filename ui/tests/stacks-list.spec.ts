/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stacks List', () => {

  test('should list available stacks', async ({ page, request }) => {
    // Seed data
    await seedCollection('test-stack-list-1', request);

    await page.goto('/stacks');
    await expect(page.getByText('test-stack-list-1', { exact: true })).toBeVisible({ timeout: 10000 });

    // Cleanup
    await cleanupCollection('test-stack-list-1', request);
  });

  test('should create a new stack', async ({ page, request }) => {
    const stackName = 'new-stack-ui-test';
    // Ensure clean state
    await cleanupCollection(stackName, request);

    await page.goto('/stacks');

    // Click Create
    await page.getByRole('button', { name: 'Create Stack' }).click();

    // Fill Form
    await page.getByLabel('Name').fill(stackName);
    await page.getByLabel('Description').fill('Created via UI test');

    // Submit
    await page.getByRole('button', { name: 'Create', exact: true }).click();

    // Verify
    await expect(page.getByText(stackName, { exact: true })).toBeVisible({ timeout: 10000 });

    // Cleanup
    await cleanupCollection(stackName, request);
  });

  test('should delete a stack', async ({ page, request }) => {
    const stackName = 'stack-to-delete-ui';
    await seedCollection(stackName, request);

    await page.goto('/stacks');
    await expect(page.getByText(stackName, { exact: true })).toBeVisible();

    // Handle dialog confirmation
    page.on('dialog', dialog => dialog.accept());

    // Find the specific card and delete button
    const card = page.locator('.group').filter({ hasText: stackName });
    await card.getByTitle('Delete Stack').click();

    // Verify it disappears
    await expect(page.getByText(stackName, { exact: true })).not.toBeVisible({ timeout: 10000 });
  });

});
