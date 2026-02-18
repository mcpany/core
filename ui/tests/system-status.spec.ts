/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('System Status', () => {
  test('should display status indicator in header', async ({ page }) => {
    await page.goto('/');

    // Wait for the indicator to appear
    const indicator = page.getByTitle('System Status');
    await expect(indicator).toBeVisible();

    // It might be "Healthy" or "Loading" initially
    await expect(indicator).toHaveText(/Healthy|Loading|Degraded/);
  });

  test('should open status sheet on click', async ({ page }) => {
    await page.goto('/');

    const indicator = page.getByTitle('System Status');
    await indicator.click();

    // Verify sheet content
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText('System Status', { exact: true })).toBeVisible();
    await expect(page.getByText('Real-time diagnostics')).toBeVisible();
  });
});
