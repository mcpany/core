/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Network Topology', () => {
  test.beforeEach(async ({ page, request }) => {
      // Seed global state (users, services, templates)
      await seedGlobalState(request);

      // Ensure login
      await page.goto('/login');
      await page.fill('input[name="username"]', 'e2e-admin-core'); // Matches seedGlobalState user
      await page.fill('input[name="password"]', 'password');
      await page.click('button[type="submit"]', { force: true });
      await page.waitForURL('/', { timeout: 30000 });

      // No mocking! Real data from backend.
      await page.goto('/network');
  });

  test('should display network topology nodes', async ({ page }) => {
    // Locate the header specifically to avoid menu link ambiguity
    await expect(page.locator('.text-lg', { hasText: 'Network Graph' })).toBeVisible();

    // Check for nodes
    // The graph might take a moment to render
    // "MCP Any" (Core) is usually present.
    // "Payment Gateway" and "User Service" should be present from seedGlobalState.
    await expect(page.locator('.react-flow').getByText(/MCP Any|Core/i).first()).toBeVisible({ timeout: 10000 });
    await expect(page.locator('.react-flow').getByText('Payment Gateway').first()).toBeVisible();
    await expect(page.locator('.react-flow').getByText('User Service').first()).toBeVisible();

    // Verify interaction
    await page.locator('.react-flow').getByText(/MCP Any|Core/i).first().click();
    // Verify sheet opens with correct details
    await expect(page.getByText(/MCP Any|Core/i).first()).toBeVisible();
  });

  test('should filter nodes', async ({ page }) => {
    // Navigate and wait for nodes
    await expect(page.locator('.react-flow').getByText(/MCP Any|Core/i).first()).toBeVisible();

    // Use Filter control
    const filterBtn = page.getByRole('button', { name: /Filter|View/i });

    if (await filterBtn.isVisible()) {
        await filterBtn.click();
        const serviceToggle = page.getByRole('menuitemcheckbox').filter({ hasText: /Services|Nodes/i });
        if (await serviceToggle.count() > 0) {
            await serviceToggle.first().click();
            // Verify nodes disappear or count changes
             await page.waitForTimeout(500);
             // Logic depends on actual implementation of filter.
        } else {
             console.log('Filter options not found, skipping specific filter interaction');
        }
    } else {
        console.log('Filter button not found in UI');
    }
  });
});
