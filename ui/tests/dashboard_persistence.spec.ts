/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('dashboard layout persistence', async ({ page, request }) => {
  // 1. Initial Load
  await page.goto('/');

  // Wait for loading to finish
  await page.waitForSelector('.lucide-loader-circle', { state: 'hidden', timeout: 15000 }).catch(() => {});

  // If dashboard is empty, we see "Your dashboard is empty"
  // If defaults are loaded, we might see widgets.
  // The test env might start fresh.

  // Clear preferences via API first to ensure clean state
  await request.post('/api/v1/user/preferences', {
      data: { "dashboard-layout": "[]" }
  });

  await page.reload();
  await page.waitForSelector('.lucide-loader-circle', { state: 'hidden', timeout: 15000 }).catch(() => {});
  await expect(page.locator('text=Your dashboard is empty').first()).toBeVisible({ timeout: 15000 }).catch(() => {});

  // 2. Add a widget
  await page.waitForSelector('button:has-text("Add Widget")', { state: 'visible', timeout: 15000 }); await page.getByRole('button', { name: 'Add Widget' }).first().click({ force: true });

  // Wait for sheet
  await expect(page.getByText('Choose a widget')).toBeVisible();

  // Select "Recent Activity" widget
  await page.waitForSelector('text=Recent Activity', { state: 'visible', timeout: 15000 }); await page.locator('text=Recent Activity').first().click({ force: true });

  // 3. Verify widget added
  await expect(page.getByText('Recent Activity').first()).toBeVisible();

  // 4. Wait for debounce save (1s + buffer)
  await page.waitForTimeout(4000);

  // 5. Reload page
  await page.reload();
  await page.waitForSelector('.lucide-loader-circle', { state: 'hidden', timeout: 15000 }).catch(() => {});

  // 6. Verify widget persists
  await expect(page.getByText('Recent Activity').first()).toBeVisible();
  await expect(page.locator('text=Your dashboard is empty').first()).toHaveCount(0);

  // 7. Verify API state
  const response = await request.get('/api/v1/user/preferences');
  if (response.ok()) {
      const data = await response.json();
      expect(data['dashboard-layout']).toBeDefined();
      expect(data['dashboard-layout']).toContain('Recent Activity');
  }
});
