/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection, cleanupServices } from './e2e/test-data';

test.describe('Dashboard Persistence', () => {
  test.beforeEach(async ({ request }) => {
    // Seed a service so we don't hit the onboarding screen
    try { await cleanupCollection('mcpany-system', request); } catch (e) { }
    try { await cleanupServices(request); } catch (e) { }
    await seedCollection('mcpany-system', request);
  });

  test.afterEach(async ({ request }) => {
    try { await cleanupCollection('mcpany-system', request); } catch (e) { }
    try { await cleanupServices(request); } catch (e) { }
  });

  test('dashboard layout persistence', async ({ page, request }) => {
    // Clear preferences via API first to ensure clean state before we even load the page
    await request.post('/api/v1/user/preferences', {
        data: { "dashboard-layout": "[]" }
    });

    // 1. Initial Load
    await page.goto('/');

    // Wait for loading to finish
    await expect(page.locator('.animate-spin')).not.toBeVisible();
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
  });
});
