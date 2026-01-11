/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Middleware Page', () => {
  test.skip('should open configuration sheet', async ({ page }) => {
    await page.goto('/middleware');

    // Find the Rate Limiter block
    // We target the block by text "Rate Limiter"
    const rateLimiter = page.locator('div').filter({ hasText: /^Rate Limiter$/ }).first();

    // We need to click the settings/gear icon inside it or the block itself if actionable.
    // Based on browser investigation, clicking the button inside worked.
    // Let's look for the button with the gear icon or just the last button in the block.
    // A reliable way might be to find the block "Rate Limiter" and find the button within it.

    // Fallback: Click the settings button specifically.
    // We'll target the common "Settings" or "Configure" label if it exists, or use a more structural selector.
    // In the browser trace, we found it by `document.querySelectorAll('button')[40]` which is brittle.
    // Better: Find the card that contains "Rate Limiter" and click the button inside it.

    const card = page.locator('.border').filter({ hasText: 'Rate Limiter' });
    const settingsBtn = card.getByRole('button').last(); // Usually the icon button is last or prominent

    // Ensure card is visible
    await expect(card).toBeVisible();

    // Click settings
    await settingsBtn.click();

    // Check for Sheet Content
    // "Configure Middleware" or the name of the middleware usually appears in the sheet title.
    await expect(page.getByText('Configure Middleware', { exact: false })).toBeVisible();
    await expect(page.getByText('Rate Limiter', { exact: false })).toBeVisible();
  });
});
