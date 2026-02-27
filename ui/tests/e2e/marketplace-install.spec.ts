/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Marketplace Installation', () => {
  test('should install GitHub service via wizard', async ({ page }) => {
    // 1. Go to Marketplace
    await page.goto('/marketplace');
    await expect(page.getByRole('heading', { name: 'Marketplace' })).toBeVisible();

    // 2. Select GitHub card and click Configure (Popular Services tab)
    // Note: The text is "Configure" now based on the UI code, but was "Instantiate" in plan.
    // The previous step implementation uses "Configure" for popular services.
    await page.getByRole('tab', { name: 'Popular' }).click();
    const githubCard = page.locator('.grid > div').filter({ hasText: 'GitHub' }).first();
    await expect(githubCard).toBeVisible();
    await githubCard.getByRole('button', { name: 'Configure' }).click();

    // 3. Verify wizard opens
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    await expect(dialog.getByText('Install Service')).toBeVisible();
    await expect(dialog.getByText('GitHub')).toBeVisible();

    // 4. Click Next to go to Configuration
    await dialog.getByRole('button', { name: 'Next' }).click();

    // 5. Fill in dummy token
    const tokenInput = dialog.getByLabel('GitHub Personal Access Token');
    await expect(tokenInput).toBeVisible();
    await tokenInput.fill('dummy_token_123');

    // 6. Click Install
    await dialog.getByRole('button', { name: 'Install' }).click();

    // 7. Wait for Success
    await expect(dialog.getByText('Installation Complete!')).toBeVisible({ timeout: 10000 });

    // 8. Go to Service
    await dialog.getByRole('button', { name: 'Go to Service' }).click();

    // 9. Verify redirection (checking URL partially as ID is random)
    await expect(page).toHaveURL(/\/upstream-services/);

    // 10. Verify we see the services list or the service itself.
    // The redirect goes to /upstream-services, so we should see the list.
    // We can check if a github service is listed.
    await expect(page.getByRole('heading', { name: 'Upstream Services' })).toBeVisible();
    // Since name is randomized (github-xxxx), we check for "github-"
    await expect(page.locator('body')).toContainText('github-');
  });
});
