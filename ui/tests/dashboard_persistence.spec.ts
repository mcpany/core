/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './e2e/test-data';

test.beforeEach(async ({ request }) => {
  // Seed the global state, creating the e2e-admin-core user
  await seedGlobalState(request);
});

test('dashboard layout persistence', async ({ page }) => {
  // 1. Login to get authenticated context
  await page.goto('/login');
  await page.fill('input[name="username"]', 'e2e-admin-core');
  await page.fill('input[name="password"]', 'password');
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL('/');

  // Wait for loading to finish
  await expect(page.locator('.animate-spin')).not.toBeVisible();

  // If dashboard is empty, we see "Your dashboard is empty"
  // If defaults are loaded, we might see widgets.
  // The test env might start fresh.

  // Clear preferences via API first to ensure clean state (must use authenticated page context)
  await page.request.post('/api/v1/user/preferences', {
      data: { "dashboard-layout": "[]" }
  });

  await page.reload();
  await expect(page.locator('.animate-spin')).not.toBeVisible();
  await expect(page.getByText('Your dashboard is empty')).toBeVisible();

  // 2. Add a widget
  await page.getByRole('button', { name: 'Add Widget' }).first().click();

  // Wait for sheet
  await expect(page.getByText('Choose a widget')).toBeVisible();

  // Select "Recent Activity" widget
  await page.getByText('Recent Activity').first().click();

  // 3. Verify widget added
  await expect(page.getByText('Recent Activity').first()).toBeVisible();

  // 4. Wait for debounce save (1s + buffer)
  await page.waitForTimeout(4000);

  // 5. Reload page
  await page.reload();
  await expect(page.locator('.animate-spin')).not.toBeVisible();

  // 6. Verify widget persists
  await expect(page.getByText('Recent Activity').first()).toBeVisible();
  await expect(page.getByText('Your dashboard is empty')).not.toBeVisible();

  // 7. Verify API state (must use authenticated page context)
  const response = await page.request.get('/api/v1/user/preferences');
  expect(response.ok()).toBeTruthy();
  const data = await response.json();
  expect(data['dashboard-layout']).toBeDefined();
  expect(data['dashboard-layout']).toContain('Recent Activity');
});
