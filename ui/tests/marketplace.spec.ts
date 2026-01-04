/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Marketplace', () => {
  test('should list popular services and allow installing them', async ({ page }) => {
    // Mock the subscriptions response if needed, but we expect real backend to provide "popular"
    // However, for pure UI test, mocking is safer.
    // Let's try real backend first if possible, but fallback to mock if flakey.
    // The previous tests mock API.

    // We will assume the backend is running and we visit /marketplace
    await page.goto('/marketplace');

    // Check for "Marketplace" header
    await expect(page.getByRole('heading', { name: 'Marketplace' })).toBeVisible();

    // Check for "Popular MCP Services"
    // It might differ if we mocked or real.
    // If real backend: "Popular MCP Services"

    // Wait for content
    await expect(page.getByText('Popular MCP Services')).toBeVisible();

    // Check for "Filesystem" service card
    await expect(page.getByText('Filesystem')).toBeVisible();

    // Click "Install" on Filesystem
    // We need to find the card that has "Filesystem" and then click "Install" inside it.
    // Or just "Install" if unique? There are multiple "Install" buttons.
    const fsCard = page.locator('.rounded-xl').filter({ hasText: 'Filesystem' });
    await fsCard.getByRole('button', { name: 'Install' }).click();

    // Should see "Service installed" toast
    await expect(page.getByText('Service installed')).toBeVisible();

    // Should now say "Installed" badge or button change
    await expect(fsCard.getByText('Installed')).toBeVisible();

    // Navigate to Services page to verify it appears
    await page.getByRole('link', { name: 'Services' }).click();

    // Should see "Filesystem" in the list
    await expect(page.getByRole('cell', { name: 'Filesystem' })).toBeVisible();

    // Verify Manage Subscriptions Tab
    await page.goto('/marketplace');
    await page.getByRole('tab', { name: 'Manage Subscriptions' }).click();
    await expect(page.getByText('Popular MCP Services')).toBeVisible();
    await expect(page.getByText('Source: internal://popular')).toBeVisible();
  });
});
