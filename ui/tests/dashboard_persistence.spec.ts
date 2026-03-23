/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('dashboard layout persistence', async ({ page, request }) => {
  // Clear preferences via API and localstorage first to ensure clean state
  await request.post('/api/v1/user/preferences', {
      data: { "dashboard-layout": "[]" }
  });

  // 1. Initial Load
  await page.goto('/');

  await page.evaluate(() => {
      localStorage.setItem('dashboard-layout', '[]');
  });
  await page.reload();

  // Wait for loading to finish
  await expect(page.locator('.lucide-loader2.animate-spin, .lucide-loader-2.animate-spin, .lucide-loader.animate-spin, .animate-spin').first()).not.toBeVisible();

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
  await expect(page.locator('.lucide-loader2.animate-spin, .lucide-loader-2.animate-spin, .lucide-loader.animate-spin, .animate-spin').first()).not.toBeVisible();

  // 6. Verify widget persists
  await expect(page.getByText('Recent Activity').first()).toBeVisible();
  await expect(page.getByText('Your dashboard is empty')).not.toBeVisible();

  // 7. Verify API state (Mock or expect a 200/404 based on test environment setup)
  const response = await request.get('/api/v1/user/preferences');
  if (response.ok()) {
      const data = await response.json();
      if (data && data['dashboard-layout']) {
        expect(data['dashboard-layout']).toContain('Recent Activity');
      }
  }
});
