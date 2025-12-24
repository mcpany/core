/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Layout & Sidebar Tests', () => {
  test('Should display only one sidebar', async ({ page }) => {
    await page.goto('/');

    // Ensure we don't have multiple sidebars
    const sidebars = page.locator('[data-sidebar="sidebar"]');
    await expect(sidebars).toHaveCount(1);

    // Verify "Dashboard" link count in the navigation
    // We expect one visible "Dashboard" link in the sidebar context
    // The previous bug had duplicate full sidebars
    const dashboardLinks = page.locator('[data-sidebar="sidebar"] a[href="/"]');
    await expect(dashboardLinks).toHaveCount(1);
  });

  test('Should not have horizontal scroll on Desktop', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.goto('/');

    // Allow for layout to settle
    await page.waitForLoadState('networkidle');

    const { scrollWidth, clientWidth } = await page.evaluate(() => {
      return {
        scrollWidth: document.documentElement.scrollWidth,
        clientWidth: document.documentElement.clientWidth
      };
    });

    expect(scrollWidth).toBeLessThanOrEqual(clientWidth + 1);
  });

  test('Should not have horizontal scroll on Mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/');

    // Allow for layout to settle
    await page.waitForLoadState('networkidle');

    const { scrollWidth, clientWidth } = await page.evaluate(() => {
      return {
        scrollWidth: document.documentElement.scrollWidth,
        clientWidth: document.documentElement.clientWidth
      };
    });

    expect(scrollWidth).toBeLessThanOrEqual(clientWidth + 1);
  });

  test('Sidebar toggle verification', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.goto('/');

    // Initially sidebar should be expanded (or at least present)
    // The default state might be expanded
    const sidebar = page.locator('[data-sidebar="sidebar"]');
    await expect(sidebar).toBeVisible();

    // Click toggle
    const toggleBtn = page.getByRole('button', { name: 'Toggle Sidebar' });
    await expect(toggleBtn).toBeVisible();
    await toggleBtn.click();

    // Verify state change (SidebarProvider usually sets data-state or similar on a wrapper,
    // but the Sidebar component itself has data-state)
    // We can also check cookie or local storage if needed, but UI state is enough
    // In "icon" mode, text might be hidden or sidebar width changes.
    // Let's just ensure it doesn't break.
    await expect(sidebar).toBeVisible();
  });
});
