/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices } from './test-data';

test.describe('Network Topology', () => {
  test.beforeEach(async ({ request }) => {
    // Clean up first to ensure fresh state
    await cleanupServices(request);
    await seedServices(request);
  });

  test.afterEach(async ({ request }) => {
    await cleanupServices(request);
  });

  test('should display network topology nodes with real data', async ({ page }) => {
    await page.goto('/network');

    // Locate the header specifically to avoid menu link ambiguity
    await expect(page.locator('.text-lg', { hasText: 'Network Graph' })).toBeVisible();

    // Check for nodes
    // The graph might take a moment to render
    // We expect "MCP Any" (Core) and the seeded services
    await expect(page.locator('.react-flow').getByText('MCP Any').first()).toBeVisible({ timeout: 15000 });
    await expect(page.locator('.react-flow').getByText('Payment Gateway').first()).toBeVisible({ timeout: 15000 });
    await expect(page.locator('.react-flow').getByText('User Service').first()).toBeVisible({ timeout: 15000 });
  });

  test('should open node details sheet', async ({ page }) => {
    await page.goto('/network');

    const coreNode = page.locator('.react-flow').getByText('MCP Any').first();
    await expect(coreNode).toBeVisible({ timeout: 15000 });
    await coreNode.click();

    // Verify sheet opens with correct details
    const sheet = page.locator('[role="dialog"]');
    await expect(sheet).toBeVisible();
    await expect(sheet.getByRole('heading', { name: /MCP Any/i })).toBeVisible();

    // Look for the type "CORE" inside the sheet specifically
    await expect(sheet.getByText(/CORE/)).toBeVisible();
  });
});
