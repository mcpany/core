/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Mobile View Verification', () => {
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
    await expect(page.locator('text=Environment Secrets')).toBeVisible();

    // Open dialog
    await page.getByRole('button', { name: 'Add Secret' }).click();

    // Check if dialog is visible
    await expect(page.getByRole('dialog')).toBeVisible();

    // Check inputs stacking (simplified check by visibility)
    await expect(page.getByLabel('Name')).toBeVisible();
    await expect(page.getByLabel('Key')).toBeVisible();
  });
});
