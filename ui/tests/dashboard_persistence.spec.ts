/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('dashboard layout persistence', async ({ page, request }) => {
  // 1. Initial Load
  await page.goto('/');

  // Wait for loading to finish
  await expect(page.locator('.animate-spin')).not.toBeVisible();

  // If dashboard is empty, we see "Your dashboard is empty"
  // If defaults are loaded, we might see widgets.
  // The test env might start fresh.

  // Clear preferences via API first to ensure clean state
  await request.post('/api/v1/user/preferences', {
      data: { "dashboard-layout": "[]" }
  });

  await page.reload();
  await expect(page.locator('.animate-spin')).not.toBeVisible();

  // We need a resilient selector because we might be on OnboardingHero OR DashboardGrid
  // Wait for either the 'Your dashboard is empty' OR the 'Connect Your First Service' link
  // If we are on onboarding hero, we need to add a dummy service to reach the dashboard.

  const welcome = page.getByText('Welcome to MCP Any');
  const dashboard = page.getByRole('heading', { name: /Dashboard/i });

  await Promise.race([
    welcome.waitFor({ state: 'visible', timeout: 5000 }).catch(() => { }),
    dashboard.waitFor({ state: 'visible', timeout: 5000 }).catch(() => { })
  ]);

  if (await welcome.isVisible()) {
    // If we're on the welcome screen, we don't have services.
    // The previous run might have cleared them. Let's create a minimal test service using the API.
    await request.post('/api/v1/services', {
        data: {
             id: "dummy-service",
             name: "dummy-service",
             version: "1.0",
             httpService: { address: "http://localhost:8080" }
        }
    });
    await page.reload();
    await expect(page.locator('.animate-spin')).not.toBeVisible();
  }

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
  await page.waitForTimeout(2000);

  // 5. Reload page
  await page.reload();
  await expect(page.locator('.animate-spin')).not.toBeVisible();

  // 6. Verify widget persists
  await expect(page.getByText('Recent Activity').first()).toBeVisible();
  await expect(page.getByText('Your dashboard is empty')).not.toBeVisible();

  // 7. Verify API state
  const response = await request.get('/api/v1/user/preferences');
  expect(response.ok()).toBeTruthy();
  const data = await response.json();
  expect(data['dashboard-layout']).toBeDefined();
  expect(data['dashboard-layout']).toContain('Recent Activity');

  // Cleanup dummy service if we created it
  if (await welcome.isVisible()) {
      await request.delete('/api/v1/services/dummy-service');
  }
});
