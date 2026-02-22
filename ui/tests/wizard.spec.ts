/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Service Installation Wizard', () => {
  test('should complete the installation wizard flow', async ({ page }) => {
    // 1. Navigate to Marketplace
    await page.goto('/marketplace');
    await expect(page.getByRole('heading', { name: 'Marketplace' })).toBeVisible();

    // 2. Click "Install" on a Community Server (Simulated by navigating to wizard directly for reliability or finding a button)
    // We'll navigate to the wizard URL to test parameter parsing
    await page.goto('/marketplace/install?name=TestService&description=TestDescription&repo=https://github.com/test/repo');

    // 3. Step 1: Identity
    await expect(page.getByLabel('Service Name')).toHaveValue('TestService');
    await expect(page.getByLabel('Description')).toHaveValue('TestDescription');
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 4. Step 2: Connection
    // Should verify heuristic
    await expect(page.getByLabel('Execution Command')).toHaveValue('npx -y repo');
    await page.getByLabel('Working Directory (Optional)').fill('/tmp');
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 5. Step 3: Configuration
    await expect(page.getByRole('button', { name: 'Secure Storage Enabled' })).toBeVisible();
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 6. Step 4: Review
    await expect(page.getByText('TestService')).toBeVisible();
    await expect(page.getByText('npx -y repo')).toBeVisible();

    // 7. Deploy
    // We mock the API call or allow it to fail if backend isn't actually creating "npx" processes
    // In E2E, we usually intercept the request or expect a failure toast if the backend validates strictly.
    // However, assuming the backend allows "npx" (it might block it for security in some envs), let's just check the button click.
    // To be safe and deterministic, we'll intercept the route.

    await page.route('**/api/v1/services', async route => {
        if (route.request().method() === 'POST') {
            await route.fulfill({ status: 200, body: JSON.stringify({ name: 'TestService' }) });
        } else {
            await route.continue();
        }
    });

    await page.getByRole('button', { name: 'Deploy Service' }).click();

    // 8. Verify Redirect
    await expect(page).toHaveURL(/\/upstream-services\/TestService/);
  });
});
