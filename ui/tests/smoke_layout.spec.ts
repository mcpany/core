/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection } from './e2e/test-data';

test('layout smoke test', async ({ page, request }) => {
  // Seed the stack data first to ensure it exists for the smoke test
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
  // Use exact match to avoid matching "No stacks found"
  await expect(page.getByRole('heading', { name: 'Stacks', exact: true })).toBeVisible({ timeout: 10000 });

  // Check for the "mcpany-system" stack
  await expect(page.locator('text=mcpany-system')).toBeVisible();

  // Navigate to Stack Detail
  await Promise.all([
    // Wait for URL matching /stacks/mcpany-system
    page.waitForURL(/\/stacks\/mcpany-system/),
    page.click('text=mcpany-system'),
  ]);
  // The header likely contains the full name "mcpany-system"
  await expect(page.locator('h2')).toContainText('mcpany-system');
  await expect(page.locator('h2')).toContainText('Stack');

  // Check Tabs
  await expect(page.locator('button[role="tab"]', { hasText: 'Overview & Status' })).toBeVisible();
  await expect(page.locator('button[role="tab"]', { hasText: 'Editor' })).toBeVisible();
});
