/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test('layout smoke test', async ({ page, request }) => {
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
  // Use exact match to differentiate from "No stacks found" empty state if present
  await expect(page.getByRole('heading', { name: 'Stacks', exact: true })).toBeVisible({ timeout: 10000 });

  // Check for the "mcpany-system" stack
  await expect(page.locator('text=mcpany-system').first()).toBeVisible();

  // Navigate to Stack Detail
  await Promise.all([
    page.waitForURL(/\/stacks\/mcpany-system/),
    page.locator('text=mcpany-system').first().click(),
  ]);
  // The seeded stack name is "mcpany-system", ID is usually same if passed to seedCollection
  await expect(page.locator('h2')).toContainText('mcpany-system');
  await expect(page.locator('h2')).toContainText('Stack');

  // Check Tabs
  await expect(page.locator('button[role="tab"]', { hasText: 'Overview & Status' })).toBeVisible();
  await expect(page.locator('button[role="tab"]', { hasText: 'Editor' })).toBeVisible();

  await cleanupCollection('mcpany-system', request);
});
