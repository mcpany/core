/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Audit Verification', () => {

  test('Verify Dashboard', async ({ page }) => {
    await page.goto('/');
    // Check for key elements from docs
    // "Total Requests", "Active Services"
    // Using simple locators
    await expect(page.locator('body')).toContainText('Total Requests');
    await expect(page.locator('body')).toContainText('Active Services');
  });

  test('Verify Services Page', async ({ page }) => {
    await page.goto('/services');
    // Check "Add Service" button
    await expect(page.getByRole('button', { name: /Add Service/i })).toBeVisible();

    // Check list headers "Name", "Type", "Status"
    await expect(page.locator('table')).toBeVisible();
    await expect(page.locator('body')).toContainText('Name');
    await expect(page.locator('body')).toContainText('Type');
  });

  test('Verify Secrets Page', async ({ page }) => {
    // Try both paths just in case
    const response = await page.goto('/settings/secrets');
    if (response?.status() === 404) {
        await page.goto('/secrets');
    }
    // Check "Add Secret" button
    await expect(page.getByRole('button', { name: /Add Secret/i })).toBeVisible();
  });

  test('Verify Logs Page', async ({ page }) => {
    await page.goto('/logs');
    // Check for search input
    await expect(page.getByPlaceholder(/Search/i)).toBeVisible();
  });

  test('Verify Playground Page', async ({ page }) => {
    await page.goto('/playground');
    // Check sidebar and main pane
    await expect(page.locator('aside')).toBeVisible();
  });

});
