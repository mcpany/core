/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stack Management', () => {
  const testStackName = 'test-mgmt-stack';
  const newStackName = 'new-ui-stack';

  test.beforeEach(async ({ request }) => {
      await cleanupCollection(testStackName, request);
      await cleanupCollection(newStackName, request);
      await seedCollection(testStackName, request);
  });

  test.afterEach(async ({ request }) => {
      await cleanupCollection(testStackName, request);
      await cleanupCollection(newStackName, request);
  });

  test('should list, create, and delete stacks', async ({ page }) => {
    await page.goto('/stacks');

    // 1. Verify seeded stack is visible
    await expect(page.getByText(testStackName)).toBeVisible();

    // 2. Create new stack
    await page.getByRole('button', { name: 'Create Stack' }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    await dialog.getByLabel('Stack Name').fill(newStackName);
    await dialog.getByRole('button', { name: 'Create Stack' }).click();

    // 3. Verify new stack appears
    // Check for the link in the table specifically to avoid toast ambiguity
    await expect(page.getByRole('link', { name: newStackName })).toBeVisible();

    // 4. Delete the seeded stack
    const row = page.getByRole('row').filter({ hasText: testStackName });

    await row.getByRole('button', { name: 'Delete' }).click();

    // Confirm deletion in AlertDialog
    const alert = page.getByRole('alertdialog');
    await expect(alert).toBeVisible();
    await alert.getByRole('button', { name: 'Delete' }).click();

    // 5. Verify it disappears
    // We check the row is gone. getByText might match the toast notification which contains the name.
    await expect(page.getByRole('row').filter({ hasText: testStackName })).not.toBeVisible();

    // New stack should still be there
    await expect(page.getByRole('link', { name: newStackName })).toBeVisible();
  });
});
