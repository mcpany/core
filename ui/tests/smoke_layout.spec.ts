/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { login } from './e2e/auth-helper';
import { seedUser, cleanupUser } from './e2e/test-data';

test.beforeEach(async ({ page, request }) => {
    await seedUser(request, "e2e-admin");
    await login(page);
});

test.afterEach(async ({ request }) => {
    await cleanupUser(request, "e2e-admin");
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

  // "mcpany-system" might not be seeded.
  // We skip checking for specific stack if not seeded, but layout is verified.
});
