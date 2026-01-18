/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Services Feature', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/upstream-services');
  });

  test('should list services, allow toggle, and manage services', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Upstream Services');

    // Verify services are listed
    await expect(page.getByText('Payment Gateway')).toBeVisible();
    await expect(page.getByText('User Service')).toBeVisible();

    // Verify Toggle exists and is interactive
    // The UI uses a Power button for toggle
    const paymentRow = page.locator('.flex-col').filter({ hasText: 'Payment Gateway' });
    const powerBtn = paymentRow.getByRole('button').first();
    await expect(powerBtn).toBeVisible();
    await powerBtn.click();

    // Register a new service via Marketplace
    await page.getByRole('link', { name: 'Add Service' }).click();
    await expect(page).toHaveURL(/.*marketplace.*/);

    // 2. Click "Create Config"
    await page.getByRole('button', { name: 'Create Config' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    // 3. Fill details (Step 1)
    const serviceName = `new-service-${Date.now()}`;
    await page.locator('#service-name').fill(serviceName);
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 4. Step 2: Parameters
    await expect(page.locator('#command')).toBeVisible();
    await page.locator('#command').fill('ls');
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 5. Step 3: Webhooks (Optional)
    await expect(page.getByText(/Webhooks & Transformers/).first()).toBeVisible();
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 6. Step 4: Auth (Optional)
    await expect(page.getByText(/Authentication/).first()).toBeVisible();
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 7. Step 5: Review
    await expect(page.getByText(/Review & Finish/).first()).toBeVisible();
    await page.getByRole('button', { name: 'Save Template', exact: true }).click();

    // After saving, it should be in Local Templates
    await page.getByRole('tab', { name: 'Local' }).click();

    // Find the specific card and instantiate it
    const card = page.locator('div.bg-card', { hasText: serviceName }).first();
    await card.getByRole('button', { name: 'Instantiate' }).click();
    await page.getByRole('button', { name: 'Create Instance' }).click();

    // It should redirect to the service page
    await expect(page).toHaveURL(new RegExp(`/upstream-services/${serviceName}-copy`), { timeout: 15000 });
    await expect(page.getByText(serviceName).first()).toBeVisible({ timeout: 10000 });
  });
});
