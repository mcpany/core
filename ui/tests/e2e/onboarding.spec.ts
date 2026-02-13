/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Onboarding Flow', () => {
  // Cleanup before each test to ensure zero state
  test.beforeEach(async ({ request }) => {
    // List services
    const listRes = await request.get('/api/v1/services');
    if (listRes.ok()) {
        const data = await listRes.json();
        const services = Array.isArray(data) ? data : (data.services || []);
        for (const s of services) {
            await request.delete(`/api/v1/services/${s.name}`);
        }
    }
  });

  test('shows onboarding when no services exist', async ({ page }) => {
    await page.goto('/');
    // Check for hero text
    await expect(page.getByText('Welcome to MCP Any')).toBeVisible({ timeout: 10000 });
    // Check for quick start cards
    await expect(page.getByText('Connect Filesystem')).toBeVisible();
    await expect(page.getByText('Connect GitHub')).toBeVisible();
  });

  test('completes quick start and redirects to dashboard', async ({ page }) => {
    await page.goto('/');

    // Click Filesystem Quick Start button
    // The card has "Connect Filesystem" title and a button "Start Setup"
    // We target the button inside the card that has "Connect Filesystem"
    // Or just the first "Start Setup" button as Filesystem is first.
    await page.getByRole('button', { name: 'Start Setup' }).first().click();

    // Dialog should open
    await expect(page.getByRole('dialog')).toBeVisible();

    // Should be in "Configure Filesystem" step (template-config)
    await expect(page.getByRole('heading', { name: 'Configure Filesystem' })).toBeVisible();

    // Fill required fields
    await page.getByLabel('Allowed Directories').fill('/tmp');

    // Click Continue
    await page.getByRole('button', { name: 'Continue' }).click();

    // Now should be in "Configure Service" step (form)
    await expect(page.getByRole('heading', { name: 'Configure Service' })).toBeVisible();

    // It should be pre-filled with "local-files" (template default name)
    await expect(page.locator('input[name="name"]')).toHaveValue('local-files');

    // IMPORTANT: Change command to a safe echo command to avoid network/installation issues during test
    await page.getByLabel('Command').fill('echo "safe mode"');

    // Click Register
    await page.getByRole('button', { name: 'Register Service' }).click();

    // Wait for Toast or Error
    await expect(page.getByText(/Service Registered|Registration Failed/)).toBeVisible();

    // Reload page to be sure
    await page.reload();

    // Should transition to Dashboard
    // "Welcome" should disappear
    await expect(page.getByText('Welcome to MCP Any')).not.toBeVisible({ timeout: 20000 });

    // Dashboard specific element should appear
    await expect(page.getByText('System Health')).toBeVisible();
  });
});
