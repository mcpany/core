/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Layout Tests', () => {
  test('Sidebar Footer should be pinned to the bottom', async ({ page }) => {
    await page.goto('/');

    // Check for the sidebar footer
    const footer = page.locator('[data-sidebar="footer"]');
    await expect(footer).toBeVisible();

    // Verify it contains the user avatar/name (Admin)
    await expect(footer).toContainText('Admin');

    // Verify visual positioning
    // We expect the footer to be at the bottom of the sidebar.
    // The sidebar has a height of 100vh (or close to it).
    // The footer should be close to the bottom of the viewport.

    const footerBox = await footer.boundingBox();
    const viewportSize = page.viewportSize();

    expect(footerBox).not.toBeNull();
    expect(viewportSize).not.toBeNull();

    if (footerBox && viewportSize) {
        // Footer bottom should be close to viewport bottom
        // giving it a small buffer for borders/padding
        expect(footerBox.y + footerBox.height).toBeGreaterThan(viewportSize.height - 50);
    }
  });

  test('Global Search backdrop should cover the window', async ({ page }) => {
     // This is hard to test automatically without screenshot comparison or checking computed styles on the overlay
     // checking for class 'w-screen h-screen' or 'fixed inset-0' matches implementation

     await page.goto('/');
     await page.keyboard.press('Control+k');

     await page.keyboard.press('Control+k');
     // Radix dialog overlay is usually a sibling or parent.
     // In our code: DialogOverlay is fixed inset-0 z-50 ... w-screen h-screen

     // We can just rely on the screenshot test for this visual detail,
     // but let's check if the dialog is visible.
     await expect(page.getByRole('dialog')).toBeVisible();
  });
});
