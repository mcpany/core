/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Layout Smoke Tests', () => {
  const stackName = 'mcpany-system';

  test.beforeEach(async ({ request }) => {
    await cleanupCollection(stackName, request);
    await seedCollection(stackName, request);
  });

  test.afterEach(async ({ request }) => {
    await cleanupCollection(stackName, request);
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

    // Check for the "mcpany-system" stack
    // Target the main title to avoid strict mode violation (subtitle also has the text)
    await expect(page.locator('.text-2xl', { hasText: stackName })).toBeVisible();

    // Navigate to Stack Detail
    await Promise.all([
      // Regex needs to match encoded or simple path
      page.waitForURL(new RegExp(`/stacks/${stackName}`)),
      page.click(`.text-2xl:has-text("${stackName}")`),
    ]);

    // Check Header
    await expect(page.locator('h2')).toContainText(stackName);
    await expect(page.locator('h2')).toContainText('Stack');

    // Check Tabs
    await expect(page.locator('button[role="tab"]', { hasText: 'Overview & Status' })).toBeVisible();
    await expect(page.locator('button[role="tab"]', { hasText: 'Editor' })).toBeVisible();
  });
});
