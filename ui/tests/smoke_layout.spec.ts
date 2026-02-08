/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Layout Smoke Test', () => {
  const STACK_NAME = 'smoke-test-stack';

  test.beforeAll(async ({ request }) => {
      await seedCollection(STACK_NAME, request);
  });

  test.afterAll(async ({ request }) => {
      await cleanupCollection(STACK_NAME, request);
  });

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

    // Check for the stack
    // Use specific selector logic
    await expect(page.locator('.grid > a').filter({ hasText: STACK_NAME })).toBeVisible();

    // Navigate to Stack Detail
    await Promise.all([
      page.waitForURL(new RegExp(`/stacks/${STACK_NAME}`)),
      page.locator('.grid > a').filter({ hasText: STACK_NAME }).click(),
    ]);
    // The ID in URL might be "mcpany-system" but the code might display it differently if seeded via Collection name
    // Collection name "mcpany-system" -> ID "mcpany-system" usually.
    await expect(page.locator('h2')).toContainText(STACK_NAME);
    await expect(page.locator('h2')).toContainText('Stack');

    // Check Tabs
    await expect(page.locator('button[role="tab"]', { hasText: 'Overview & Status' })).toBeVisible();
    await expect(page.locator('button[role="tab"]', { hasText: 'Editor' })).toBeVisible();
  });
});
