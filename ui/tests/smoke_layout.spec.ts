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

  // Check for the "mcpany-system" stack OR empty state
  const systemStack = page.locator('text=mcpany-system');
  const emptyState = page.getByText(/No stacks found/i);

  // Wait until one of them is visible to avoid race conditions
  await expect(async () => {
      const isStackVisible = await systemStack.isVisible();
      const isEmptyVisible = await emptyState.isVisible();
      expect(isStackVisible || isEmptyVisible).toBe(true);
  }).toPass({ timeout: 15000 });

  if (await systemStack.isVisible()) {
    // Navigate to Stack Detail
    await Promise.all([
      page.waitForURL(/\/stacks\/(mcpany-system|system)/),
      systemStack.click(),
    ]);
    // With new layout, header might be h1 or h2
    await expect(page.getByRole('heading', { name: /system/i })).toBeVisible();

    // Check Tabs - StackEditor has Editor, Code, Visualizer
    await expect(page.getByRole('tab', { name: 'Editor' })).toBeVisible();
  } else {
    // If no stack, we should see "No stacks found" or similar
    await expect(emptyState).toBeVisible();
  }
});
