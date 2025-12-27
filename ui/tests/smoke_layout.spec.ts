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
  await expect(page.locator('a[href="/services"]').first()).toBeVisible();
  await expect(page.locator('a[href="/settings"]').first()).toBeVisible();

  // Navigate to Stacks
  await page.locator('a[href="/stacks"]').first().click();
  await page.waitForURL('**/stacks');
  await expect(page.getByRole('heading', { name: 'Stacks' })).toBeVisible({ timeout: 10000 });

  // Check for the "mcpany-system" stack
  await expect(page.locator('text=mcpany-system')).toBeVisible();

  // Navigate to Stack Detail
  await page.click('text=mcpany-system');
  await expect(page.locator('h1')).toContainText('Stack: system');

  // Check Tabs
  await expect(page.locator('button[role="tab"]', { hasText: 'Services' })).toBeVisible();
  await expect(page.locator('button[role="tab"]', { hasText: 'Editor' })).toBeVisible();
});
