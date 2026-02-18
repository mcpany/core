/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Smart Config Wizard', () => {
  const serviceName = 'Test Postgres Wizard';

  test.beforeEach(async ({ request }) => {
    // Cleanup: Ensure service does not exist
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});
  });

  test.afterEach(async ({ request }) => {
    // Cleanup
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});
  });

  test('should verify the full wizard flow for PostgreSQL template', async ({ page }) => {
    page.on('console', msg => console.log('BROWSER LOG:', msg.text()));
    page.on('pageerror', err => console.log('BROWSER ERROR:', err));

    // 1. Navigate to Marketplace
    await page.goto('/marketplace');
    await expect(page.getByRole('heading', { name: 'Marketplace' })).toBeVisible();

    // 2. Open Wizard
    const createButton = page.getByRole('button', { name: 'Create Config' });

    // Ensure button is ready and attached
    await expect(createButton).toBeVisible();
    await expect(createButton).toBeEnabled();

    // Force click to ensure it works even if toasts overlay
    await createButton.click({ force: true });

    // Check if dialog opened with increased timeout
    try {
        // Try alternate locator strategy if heading isn't immediately found
        const dialog = page.locator('div[role="dialog"]');
        await expect(dialog).toBeVisible({ timeout: 15000 });
        await expect(page.getByText('Create Upstream Service Config')).toBeVisible({ timeout: 5000 });
    } catch (e) {
        console.log('Dialog not visible. Retrying click...');
        // Retry click once if dialog didn't open (hydration race condition mitigation)
        await createButton.click({ force: true });
        await expect(page.locator('div[role="dialog"]')).toBeVisible({ timeout: 15000 });
    }

    // 3. Select Template
    // The Select trigger usually has the current value or placeholder
    const templateTrigger = page.locator('button[role="combobox"]');
    await templateTrigger.click();
    await page.getByRole('option', { name: 'PostgreSQL' }).click();

    // 4. Set Name
    await page.getByLabel('Service Name').fill(serviceName);

    // 5. Click Next (Step 1 -> 2)
    // Use exact match to avoid matching Next.js dev tools button
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 6. Fill Parameters (Schema Driven)
    // Wait for the schema form to render
    await expect(page.getByText('Connection URL')).toBeVisible();
    await page.getByLabel('Connection URL').fill('postgresql://test:test@localhost:5432/testdb');

    // 7. Click Next (Step 2 -> 3)
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 8. Click Next (Step 3 -> 4)
    await expect(page.getByText('3. Webhooks & Transformers')).toBeVisible();
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 9. Click Next (Step 4 -> 5)
    await expect(page.getByText('4. Authentication')).toBeVisible();
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 10. Review & Finish
    await expect(page.getByText('Review & Finish')).toBeVisible();
    // Use regex to be more flexible with formatting
    await expect(page.getByText(/"name":\s*"Test Postgres Wizard"/)).toBeVisible();

    // Click Finish
    await page.getByRole('button', { name: 'Finish & Save to Local Marketplace' }).click();

    // 11. Verify Success Toast
    // Increase timeout for backend response
    await expect(page.getByText('Config Saved')).toBeVisible({ timeout: 20000 });

    // 12. Verify in "Local Templates" tab
    // Wait for the modal to close or toast to disappear to prevent click interception
    await page.getByRole('tab', { name: 'Local' }).click({ force: true });
    await expect(page.getByRole('heading', { name: serviceName })).toBeVisible({ timeout: 20000 });
  });
});
