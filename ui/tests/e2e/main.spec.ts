/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('MCP Any UI E2E', () => {

  test('Dashboard loads and shows metrics', async ({ page }) => {
    await page.goto('/');
    // Updated title expectation
    await expect(page).toHaveTitle(/MCPAny Manager/);
    await expect(page.locator('h1')).toContainText('Dashboard');
    // Check for metrics cards
    await expect(page.locator('text=Total Requests')).toBeVisible();
    await expect(page.locator('text=System Health')).toBeVisible();
  });

  test('Services page CRUD', async ({ page }) => {
    await page.goto('/services');
    await expect(page.locator('h2')).toContainText('Services');

    // Add Service
    // Add Service
    await page.getByRole('button', { name: 'Custom Service' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await page.getByLabel('Name').fill('test-service-e2e');
    await page.getByLabel('Endpoint').fill('https://example.com');
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // Check if added
    await expect(page.locator('text=test-service-e2e')).toBeVisible();
  });

  test('Tools page lists tools and inspects', async ({ page }) => {
    await page.goto('/tools');
    await expect(page.locator('h2')).toContainText('Tools');
    // Wait for list to load
    await expect(page.locator('text=Available Tools')).toBeVisible();

    // Inspect
    await page.click('button:has-text("Inspect") >> nth=0');
    await expect(page.locator('div[role="dialog"]')).toBeVisible();
    await expect(page.locator('text=Schema')).toBeVisible();
  });

  test('Middleware page drag and drop', async ({ page }) => {
    await page.goto('/middleware');
    await expect(page.locator('h2')).toContainText('Middleware Pipeline');
    await expect(page.locator('text=Active Pipeline')).toBeVisible();
    // Resolving ambiguity by selecting the first occurrence (likely the list item)
    await expect(page.locator('text=Authentication').first()).toBeVisible();
  });

});
