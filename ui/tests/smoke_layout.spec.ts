/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection } from './e2e/test-data';

test('layout smoke test', async ({ page, request }) => {
  // Ensure the default system stack exists
  await seedCollection('mcpany-system', request);

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

  // Check for the "mcpany-system" stack
  await expect(page.locator('text=mcpany-system')).toBeVisible();

  // Navigate to Stack Detail
  await Promise.all([
    page.waitForURL(/\/stacks\/mcpany-system/),
    page.click('text=mcpany-system'),
  ]);
  // The header might be just "mcpany-system" or include "Stack" badge
  await expect(page.locator('h2')).toContainText('mcpany-system');

  // Check Tabs
  await expect(page.locator('button[role="tab"]', { hasText: 'Overview & Status' })).toBeVisible();
  await expect(page.locator('button[role="tab"]', { hasText: 'Editor' })).toBeVisible();
});
