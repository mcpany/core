/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Stacks Management', () => {
  const stackName = 'e2e-test-stack';

  test.beforeEach(async ({ request }) => {
      // Ensure clean state
      try {
          await request.delete(`/api/v1/collections/${stackName}`);
      } catch (e) {
          // Ignore if not exists
      }
  });

  test.afterEach(async ({ request }) => {
      // Cleanup
      try {
          await request.delete(`/api/v1/collections/${stackName}`);
      } catch (e) {
          // Ignore
      }
  });

  test('should create, list, and delete a stack', async ({ page }) => {
    // 1. Navigate to Stacks page
    await page.goto('/stacks');

    // 2. Click Create Stack
    await page.getByRole('button', { name: 'Create Stack' }).first().click();

    // 3. Fill Dialog
    await expect(page.getByRole('dialog')).toBeVisible();
    await page.getByLabel('Stack Name').fill(stackName);
    await page.getByRole('button', { name: 'Create' }).click();

    // 4. Verify Redirection to Editor
    await expect(page).toHaveURL(`/stacks/${stackName}`);
    await expect(page.getByRole('heading', { name: stackName })).toBeVisible();

    // 5. Navigate back to List
    await page.goto('/stacks');

    // 6. Verify Stack is Listed
    await expect(page.getByText(stackName, { exact: true })).toBeVisible();

    // 7. Delete Stack
    // We need to hover to see the delete button if using the card style, or just force click if hidden?
    // The implementation has "group-hover:opacity-100".
    // Playwright can click invisible elements if force:true, or we can hover.
    const card = page.locator('.group').filter({ hasText: stackName });
    await card.hover();

    await card.getByTitle('Delete Stack').click();

    // Handle Shadcn Dialog
    await expect(page.getByRole('dialog')).toBeVisible();
    await page.getByRole('button', { name: 'Delete Stack' }).click();

    // 8. Verify Stack is Gone
    // Target specific card element to avoid matching toast notification
    const stackCard = page.locator('.text-2xl.font-bold').filter({ hasText: stackName });
    await expect(stackCard).not.toBeVisible();
  });
});
