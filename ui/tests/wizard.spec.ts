/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Install Wizard', () => {
  test('should navigate through wizard steps', async ({ page }) => {
    // Generate unique name
    const serviceName = `WizardTestService-${Date.now()}`;

    // Navigate to wizard
    await page.goto(`/marketplace/install?name=${serviceName}&repo=https://github.com/example/repo`);

    // Step 1: Identity
    await expect(page.getByLabel('Service Name')).toHaveValue(serviceName);
    await expect(page.getByLabel('Source Repository')).toHaveValue('https://github.com/example/repo');
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // Step 2: Connection
    await page.getByLabel('Command').fill('echo "hello wizard"');
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // Step 3: Config
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // Step 4: Auth
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // Step 5: Review
    await expect(page.getByText(serviceName)).toBeVisible();
    await expect(page.getByText('echo "hello wizard"')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Deploy' })).toBeVisible();
  });
});
