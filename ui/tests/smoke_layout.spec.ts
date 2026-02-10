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

  // Check for ANY stack OR empty state
  // This is more robust than checking for "mcpany-system" which might vary in E2E environments
  const stackLinks = page.locator('a[href^="/stacks/"]');
  const emptyState = page.getByText(/No stacks found/i);

  // Wait for the list to load (either empty state or at least one stack)
  await expect(async () => {
      const stackCount = await stackLinks.count();
      const isEmptyVisible = await emptyState.isVisible();
      expect(stackCount > 0 || isEmptyVisible).toBe(true);
  }).toPass({ timeout: 30000 });

  const count = await stackLinks.count();
  if (count > 0) {
    // Navigate to the first available stack
    const firstStack = stackLinks.first();
    const href = await firstStack.getAttribute('href');
    const stackName = href?.split('/').pop() || 'system';

    await Promise.all([
      page.waitForURL(new RegExp(`/stacks/${stackName}`)),
      firstStack.click(),
    ]);

    // Header should contain the stack name
    await expect(page.getByRole('heading', { name: stackName, exact: false })).toBeVisible();

    // Check Tabs - StackEditor has Editor, Code, Visualizer
    await expect(page.getByRole('tab', { name: 'Editor' })).toBeVisible();
  } else {
    // If no stack, we should see "No stacks found" or similar
    await expect(emptyState).toBeVisible();
  }
});
