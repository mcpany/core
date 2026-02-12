/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Onboarding Flow', () => {
  // Reset state before tests if possible, or assume fresh env
  // Ideally call API to delete all services.

  test('should show onboarding view when no services exist', async ({ page }) => {
    // Note: This test requires a clean backend state (0 services).
    await page.goto('/');

    // Check for hero text
    await expect(page.getByText('Welcome to MCP Any')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Connect your first server')).toBeVisible();

    // Check for cards
    await expect(page.getByText('Memory Server')).toBeVisible();
    await expect(page.getByText('Marketplace')).toBeVisible();
    await expect(page.getByText('Connect Manually')).toBeVisible();
  });

  test('should register memory server and redirect to dashboard', async ({ page }) => {
    await page.goto('/');

    // Click install on Memory Server card
    // We target the specific button inside the first card or by text
    const installBtn = page.getByRole('button', { name: 'One-Click Install' });
    // Ensure we are clicking the one for Memory if there are multiple (currently only one has this text)
    await installBtn.click();

    // Wait for success toast
    await expect(page.getByText('Service Installed')).toBeVisible({ timeout: 15000 });

    // Wait for dashboard grid to appear.
    // The DashboardGrid renders widgets.
    // We expect "System Status" or "Metrics" or just absence of Onboarding text.
    await expect(page.getByText('Welcome to MCP Any')).not.toBeVisible({ timeout: 10000 });

    // Verify a service is now listed in the services page
    await page.goto('/upstream-services');
    await expect(page.getByText('memory')).toBeVisible();
  });
});
