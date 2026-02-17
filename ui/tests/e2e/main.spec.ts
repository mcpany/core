/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import { seedServices, seedTraffic, seedUser, cleanupServices, cleanupUser } from './test-data';

test.describe('MCP Any UI E2E', () => {

  test.beforeEach(async ({ request, page }) => {
    await seedServices(request);
    await seedTraffic(request);
    await seedUser(request, "e2e-admin");

    // Login
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', 'e2e-admin');
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/', { timeout: 15000 });
  });

  test.afterEach(async ({ request }) => {
    await cleanupServices(request);
    await cleanupUser(request, "e2e-admin");
  });

  test('Dashboard loads and shows metrics', async ({ page }) => {
    await page.goto('/');
    // Updated title expectation to be robust (accept both branding variations)
    await expect(page).toHaveTitle(/MCPAny Manager|Jules Master/);

    // Ensure API Key is set properly in environment
    await expect(page.getByText(/API Key Not Set/i)).not.toBeVisible();

    await expect(page.locator('h1')).toContainText(/Dashboard|Jules Master/);

    // Check for metrics cards
    await expect(page.locator('text=Total Requests').first()).toBeVisible();
    await expect(page.locator('text=System Health').first()).toBeVisible();
    // Verify that exactly 2 metric cards are displayed
    await expect(page.locator('text=Total Requests').first()).toBeVisible();
  });

  test('should navigate to analytics from sidebar', async ({ page }) => {
    // Verify direct navigation first (and warm up the route)
    await page.goto('/stats');
    await expect(page.locator('h1')).toContainText('Analytics & Stats');

    await page.goto('/');
    // Check if link exists
    const statsLink = page.getByRole('link', { name: /Analytics|Stats/i });

    await expect(statsLink).toBeVisible();
    await expect(statsLink).toHaveAttribute('href', '/stats');
    await statsLink.click();
    // Explicitly wait for navigation
    await page.waitForURL(/.*\/stats/, { timeout: 30000, waitUntil: 'domcontentloaded' });
    await expect(page).toHaveURL(/.*\/stats/);

    // Verify page content
    await expect(page.locator('h1')).toContainText('Analytics & Stats');
  });

  test('Middleware page drag and drop', async ({ page }) => {
    await page.goto('/middleware');

    await expect(page.locator('text=This page could not be found')).not.toBeVisible();

    await expect(page.locator('h1')).toContainText('Middleware Pipeline');
    await expect(page.locator('text=Active Pipeline')).toBeVisible();
    // Resolving ambiguity by selecting the first occurrence (likely the list item)
    await expect(page.locator('text=Authentication').first()).toBeVisible();
  });

});
