/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test('dashboard loads and displays metrics', async ({ page }) => {
  await page.goto('/');

  // Check for title
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();

  // Check for metrics
  await expect(page.getByText('Total Requests')).toBeVisible();
  await expect(page.getByText('Active Services')).toBeVisible();

  // Check for service health widget
  await expect(page.getByText('Service Health')).toBeVisible();
  await expect(page.getByText('Payment Gateway')).toBeVisible();
});
