/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test('settings page loads', async ({ page }) => {
  await page.goto('/settings');
  await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible();

  // Check tabs
  await expect(page.getByRole('tab', { name: 'Profiles' })).toBeVisible();
  await expect(page.getByRole('tab', { name: 'Webhooks' })).toBeVisible();
  await expect(page.getByRole('tab', { name: 'Middleware' })).toBeVisible();
});
