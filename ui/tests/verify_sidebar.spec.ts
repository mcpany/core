/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import fs from 'fs';
import path from 'path';
import { seedUser, cleanupUser } from './e2e/test-data';

test('verify sidebar navigation', async ({ request, page }) => {
  await seedUser(request, "e2e-sidebar-admin");

  // Login
  await page.goto('/login');
  await page.waitForLoadState('networkidle');
  await page.fill('input[name="username"]', 'e2e-sidebar-admin');
  await page.fill('input[name="password"]', 'password');
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL('/', { timeout: 15000 });

  // Go to homepage
  await page.goto('/');

  // Check if sidebar is visible
  await expect(page.locator('button[data-sidebar="trigger"]')).toBeVisible();

  // Check if "Platform" group exists
  await expect(page.getByText('Platform')).toBeVisible();

  // Check links
  await expect(page.getByRole('link', { name: 'Dashboard' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Network Graph' })).toBeVisible();

  // Take screenshot
  const date = new Date().toISOString().split('T')[0];
  const dir = path.join('test-results', 'audit', 'ui', date);
  if (!fs.existsSync(dir)){
      fs.mkdirSync(dir, { recursive: true });
  }
  await page.screenshot({ path: path.join(dir, 'unified_navigation_system.png') });
});
