/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test('settings page loads', async ({ page }) => {
  await page.goto('/settings');
  await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible();

  // Updated expectations: Check for Card titles
  // CardTitle uses a 'div' by default in the shadcn implementation provided.
  // We can just search for the text.

  await expect(page.getByText('Profiles', { exact: true })).toBeVisible();
  await expect(page.getByText('Webhooks', { exact: true })).toBeVisible();
  await expect(page.getByText('Middleware', { exact: true })).toBeVisible();
});
