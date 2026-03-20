/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './e2e/test-data';

test.describe('Dashboard Widget Gallery', () => {
    test.beforeEach(async ({ request, page }) => {
        // Use the common seeding logic that works reliably across the test suite
        await seedGlobalState(request);

        // Login UI with the user created by seedGlobalState
        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-admin-core');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        // seedGlobalState manages its own cleanup implicitly by overwriting or it's cleaned up globally
    });

    test('should show widget gallery and render Service Gallery with seeded data', async ({ page }) => {
        await page.goto('/');

  // Check if "Add Widget" button exists
  await expect(page.getByRole('button', { name: 'Add Widget' })).toBeVisible();

  // Open the sheet
  await page.getByRole('button', { name: 'Add Widget' }).click();

  // Check if gallery is visible
  await expect(page.getByText('Choose a widget to add')).toBeVisible();

  // Check for the new Service Gallery widget in the gallery
  await expect(page.locator('.grid > div').filter({ hasText: 'Service Gallery' })).toBeVisible();

  // Add a "Service Health" widget.
  // In the gallery it is called "Service Health".
  // Use filter to be specific.
  const galleryItem = page.locator('.grid > div').filter({ hasText: 'Service Health' }).first();
  await galleryItem.click();

  // Verify we have widgets on the dashboard.
  // The ServiceHealthWidget renders with title "System Health".
  // Since we might have multiple, let's just check that at least one is visible.
  await expect(page.getByText('System Health').first()).toBeVisible();

  // Check if the Service Gallery widget is rendered on the dashboard natively
  // It is part of the default widgets now so it should be visible after page load
  // And it should show the seeded services (Payment Gateway, Echo Service) fetched from the backend db
  const activeServicesWidget = page.locator('.group\\/widget', { hasText: 'Connected Services' }).first();
  await expect(activeServicesWidget).toBeVisible({ timeout: 10000 });
  await expect(activeServicesWidget.getByText('Payment Gateway')).toBeVisible();

  // Optionally check count if we started with known state, but default layout might change.
    });
});
