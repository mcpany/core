/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('setup wizard flow', async ({ page }) => {
  // Go to setup page directly
  await page.goto('/setup');

  // Step 1: Welcome
  await expect(page.getByText('Welcome to MCP Any')).toBeVisible();
  await page.getByRole('button', { name: 'Get Started' }).click();

  // Step 2: Template Selector
  await expect(page.getByText('Choose a Starter Template')).toBeVisible();

  // Select "Time" template (search for it to be safe)
  await page.getByPlaceholder('Search templates...').fill('Time');
  // Click on the card that contains "Time" - be specific to avoid matching "Time" in description or category
  await page.locator('.font-semibold').filter({ hasText: 'Time' }).click();

  // Step 3: Config
  // Time template has no fields, but ConfigStep should still show "Configure Service" and "Continue"
  await expect(page.getByText('Configure Service')).toBeVisible();

  // Submit
  await page.getByRole('button', { name: 'Continue' }).click();

  // Step 4: Success
  // Wait for success message (might take a moment for backend)
  await expect(page.getByText('You are All Set!')).toBeVisible({ timeout: 10000 });

  // Go to Dashboard
  await page.getByRole('button', { name: 'Go to Dashboard' }).click();

  // Verify we are on dashboard
  await expect(page).toHaveURL('/');

  // Verify the "Welcome" CTA on dashboard is gone (meaning services > 0)
  // We expect "Dashboard" heading instead
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
  await expect(page.getByText('Run Setup Wizard')).not.toBeVisible();
});
