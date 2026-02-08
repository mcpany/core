/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('layout smoke test', async ({ page }) => {
  await page.goto('/');

  // Check for Sidebar
  const sidebar = page.locator('text=MCP Any').first();
  await expect(sidebar).toBeVisible();

  // Check for Sidebar links
  await expect(page.locator('a[href="/stacks"]').first()).toBeVisible();
  await expect(page.locator('a[href="/upstream-services"]').first()).toBeVisible();
  await expect(page.locator('a[href="/settings"]').first()).toBeVisible();

  // Navigate to Stacks
  await page.locator('a[href="/stacks"]').first().click();
  await page.waitForURL('**/stacks');
  await expect(page.getByRole('heading', { name: 'Stacks' })).toBeVisible({ timeout: 10000 });

  // Handle empty state gracefully
  const hasEmptyState = await page.getByText('No stacks found').isVisible();
  if (!hasEmptyState) {
      // If not empty, we assume at least one stack card is visible
      // We don't check for "mcpany-system" specifically anymore as it's not hardcoded
      await expect(page.locator('.grid > a').first()).toBeVisible();
  } else {
      // If empty, verify the empty state message
      await expect(page.getByText('Create your first stack')).toBeVisible();
  }
});
