/**
 * Copyright 2026 Author(s) of MCP Any
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

test('verify sidebar navigation', async ({ page }) => {
  // Go to homepage
  await page.goto('/');

  // Check if sidebar is visible (might be collapsed or toggle button)
  const trigger = page.locator('button[data-sidebar="trigger"]');
  // It might be hidden on desktop if sidebar is always open?
  // Let's just check for nav links which proves sidebar content is loaded.

  // Check links
  await expect(page.getByRole('link', { name: 'Dashboard' }).first()).toBeVisible();
  await expect(page.getByRole('link', { name: 'Network Graph' }).first()).toBeVisible();

  // Take screenshot if needed
  // await page.screenshot({ path: `.audit/ui/${new Date().toISOString().split('T')[0]}/unified_navigation_system.png` });
});
