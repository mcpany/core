/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './e2e/test-data';

test('dashboard layout persistence', async ({ page, request }) => {
  // 0. Seed global state (Database Seeding constraint)
  await seedGlobalState(request);

  // 1. Login to get authenticated session
  await page.goto('/login');
  await page.fill('input[name="username"]', 'e2e-admin-core');
  await page.fill('input[name="password"]', 'password');
  await page.click('button[type="submit"]');

  // Wait for successful login and redirect to home
  await page.waitForURL('**/*');

  // Create an authenticated request context
  const storageState = await page.context().storageState();
  const authRequest = await request.newContext({
      storageState: storageState
  });

  // Wait for loading to finish
  await expect(page.locator('.animate-spin')).not.toBeVisible();

  // Clear preferences via API first to ensure clean state
  await authRequest.post('/api/v1/user/preferences', {
      data: { "dashboard-layout": "[]" },
      headers: {
          'Content-Type': 'application/json'
      }
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

  // 7. Verify API state
  const response = await authRequest.get('/api/v1/user/preferences');
  expect(response.ok()).toBeTruthy();
  const data = await response.json();
  expect(data['dashboard-layout']).toBeDefined();
  expect(data['dashboard-layout']).toContain('Recent Activity');
});
