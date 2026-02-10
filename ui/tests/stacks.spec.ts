/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('create and edit stack', async ({ page }) => {
  // 1. Navigate to Stacks page
  await page.goto('/stacks');

  // 2. Click Create Stack
  await page.getByRole('button', { name: 'Create Stack' }).click();
  await expect(page).toHaveURL(/\/stacks\/new/);

  // 3. Wait for content to load
  await expect(page.locator('.monaco-editor')).toBeVisible();
  await page.waitForTimeout(1000);

  // Click Deploy Stack
  await page.getByRole('button', { name: 'Deploy Stack' }).click();

  // 4. Verify redirection to /stacks/new-stack
  await expect(page).toHaveURL(/\/stacks\/new-stack/, { timeout: 20000 });

  // 5. Verify success toast
  await expect(page.getByText('Stack Saved')).toBeVisible();

  // 6. Navigate back to list to verify persistence
  await page.goto('/stacks');

  // 7. Verify new-stack is in the list
  await expect(page.getByText('new-stack')).toBeVisible();
});
