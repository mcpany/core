/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Wizard E2E', () => {
  // Use a unique name to avoid conflicts if backend state persists
  const serviceName = `postgres-wizard-${Date.now()}`;

  test('creates a PostgreSQL service using the wizard', async ({ page }) => {
    // 1. Navigate to Marketplace
    await page.goto('/marketplace');

    // 2. Open Wizard
    await page.waitForSelector('button:has-text("Create Config")', { state: 'visible' });
    await page.click('button:has-text("Create Config")');
    await expect(page.getByText('Create Upstream Service Config')).toBeVisible();

    // 3. Select Template
    // The Select Trigger has id="service-template"
    await page.waitForSelector('button#service-template', { state: 'visible' });
    await page.click('button#service-template');

    // Wait for dropdown content - using role="option" or "item" depending on Radix UI version/impl
    // We try to be robust.
    // In service-registry, Postgres name is "PostgreSQL"
    // In StepServiceType, it renders SelectItem with value=id, children=name
    // SelectItem usually has role="option".
    const postgresOption = page.locator('div[role="option"]:has-text("PostgreSQL")').or(page.locator('div[role="item"]:has-text("PostgreSQL")'));
    await postgresOption.first().click();

    // 4. Fill Service Name
    await page.fill('input#service-name', serviceName);

    // 5. Next Step (Parameters)
    await page.click('button:has-text("Next")');

    // 6. Verify Schema Form
    // Should see "Service Configuration" and "Connection URL"
    await expect(page.getByText('Service Configuration')).toBeVisible();
    await expect(page.getByLabel('Connection URL')).toBeVisible();

    // 7. Update Parameter
    // The schema defines POSTGRES_URL with format: password, so it should be type="password"
    await page.fill('input[type="password"]', 'postgresql://user:pass@db:5432/testdb');

    // 8. Next Step (Webhooks - Skip)
    await page.click('button:has-text("Next")');
    await expect(page.getByText('3. Webhooks & Transformers')).toBeVisible();

    // 9. Next Step (Auth - Skip)
    await page.click('button:has-text("Next")');
    await expect(page.getByText('4. Authentication')).toBeVisible();

    await page.click('button:has-text("Next")');

    // 10. Review & Finish
    await expect(page.getByText('5. Review & Finish')).toBeVisible();

    // 11. Deploy
    // Button text changed to "Deploy Service"
    await page.click('button:has-text("Deploy Service")');

    // 12. Verify Toast
    // Expect "Service Deployed"
    await expect(page.getByText('Service Deployed')).toBeVisible();
  });
});
