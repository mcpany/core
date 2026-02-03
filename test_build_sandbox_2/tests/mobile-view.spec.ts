/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Mobile View Verification', () => {
  test('should toggle sidebar', async ({ page }) => {
    await page.goto('/');

    // Verify Sidebar is hidden initially (outside viewport or explicitly hidden class)
    // The sidebar usually has a specific nav role or class.
    // In mobile, checking that the "Dashboard" *text* in the sidebar is not visible or obscured is a good proxy.
    const sidebarLink = page.getByRole('link', { name: 'Marketplace' });
    await expect(sidebarLink).toBeHidden();

    // Click Hamburger Menu
    // Usually top-left.
    const menuBtn = page.getByRole('button').first(); // Often the first button in header
    // Or better, look for a likely icon label if accessible, or class.
    // Browser trace showed it was just a button in the header.
    await menuBtn.click();

    // Verify Sidebar is now visible
    await expect(sidebarLink).toBeVisible();

    // Click outside or close to dismiss (optional, but good practice)
    // await page.locator('.fixed.inset-0').click(); // Overlay
  });

  test.use({ viewport: { width: 375, height: 667 } }); // iPhone SE

  test('should render Network Graph controls in mobile mode', async ({ page }) => {
    await page.goto('/network');

    // In mobile mode, we render:
    // <span className={cn(isControlsExpanded ? "block" : "hidden sm:block")}>Network Graph</span>

    // Initially isControlsExpanded is false on mobile. So "Network Graph" text is HIDDEN.
    // The card header is still there, but only the icon is visible (Activity icon) and the badges.

    // Let's verify that the Card is visible.
    // Use a locator that finds the card by class or by a visible element inside it.
    // The "0 Nodes" badge should be visible.
    const nodesBadge = page.getByText('Nodes', { exact: false });
    await expect(nodesBadge).toBeVisible();

    // Check if filters are hidden initially
    const filters = page.getByText('Filters');
    await expect(filters).not.toBeVisible();

    // Click to expand. We can click the card header.
    // Since "Network Graph" text is hidden, we can click the Nodes badge parent or just the card itself.
    // Let's click the card container.
    const card = page.locator('.pointer-events-auto').first();
    await card.click();

    // Now "Filters" should be visible
    await expect(filters).toBeVisible();

    // And "Network Graph" text should be visible now
    const title = page.getByText('Network Graph');
    await expect(title).toBeVisible();
  });

  test('should render Log Stream with mobile layout', async ({ page }) => {
    await page.goto('/logs');
    // Check if pause button is visible in header (mobile only)
    // The mobile pause button is in a div with md:hidden.
    const mobilePauseContainer = page.locator('.md\\:hidden');
    await expect(mobilePauseContainer).toBeVisible();

    // Use a more specific locator for the title
    const title = page.locator('h1', { hasText: 'Live Logs' });
    await expect(title).toBeVisible();
  });

  test('should render Secret Manager with mobile layout', async ({ page }) => {
    await page.goto('/secrets');
    // Check if table renders
    await expect(page.locator('text=API Keys & Secrets')).toBeVisible();

    // Open dialog
    await page.getByRole('button', { name: 'Add Secret' }).click();

    // Check if dialog is visible
    await expect(page.getByRole('dialog')).toBeVisible();

    // Check inputs stacking (simplified check by visibility)
    await expect(page.getByLabel('Friendly Name')).toBeVisible();
    await expect(page.getByLabel('Key Name (Env Var)')).toBeVisible();
  });
});
