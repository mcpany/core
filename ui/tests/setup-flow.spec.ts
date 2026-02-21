/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Setup Wizard Flow', () => {
  test('complete setup wizard with Weather service', async ({ page }) => {
    // 1. Navigate to Setup Wizard
    await page.goto('/setup');

    // 2. Welcome Step
    await expect(page.getByText('Welcome to MCP Any')).toBeVisible();
    await page.getByRole('button', { name: 'Get Started' }).click();

    // 3. Template Selection Step
    await expect(page.getByText('Choose a Starter Template')).toBeVisible();

    // Select "Weather (wttr.in)" template
    await page.getByText('Get real-time weather information via wttr.in.').click();

    // 4. Configuration Step
    // Weather template has no fields, so we should just see the "Continue" button
    await expect(page.getByRole('button', { name: 'Continue' })).toBeVisible();
    await page.getByRole('button', { name: 'Continue' }).click();

    // Check for error toast
    const errorToast = page.locator('.text-destructive');
    if (await errorToast.isVisible()) {
        console.log('Error Toast:', await errorToast.textContent());
    }

    // 5. Success Step
    await expect(page.getByText("You're All Set!")).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Your service Weather (wttr.in) is now active')).toBeVisible();

    // 6. Navigation
    await page.getByRole('button', { name: 'Go to Dashboard' }).click();
    await expect(page).toHaveURL(/\/$/);
  });
});
