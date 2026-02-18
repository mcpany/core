/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Wizard E2E', () => {
  // Use a unique name to avoid conflicts if backend state persists
  const serviceName = `postgres-wizard-${Date.now()}`;

  test('creates a PostgreSQL service using the wizard', async ({ page, request }) => {
    // Mock the backend API for templates since it might not be implemented in the test env backend
    await page.route('**/api/v1/templates', async route => {
      if (route.request().method() === 'POST') {
        const json = { name: serviceName, service_config: { command_line_service: { env: { POSTGRES_URL: { plain_text: 'postgresql://user:pass@db:5432/testdb' } } } } };
        await route.fulfill({ json });
      } else {
        await route.continue();
      }
    });

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
    // Wait for dropdown content
    await page.waitForSelector('div[role="item"]', { state: 'visible', timeout: 5000 }).catch(() => page.waitForSelector('div[role="option"]'));
    await page.click('text="PostgreSQL"');

    // 4. Fill Service Name
    await page.fill('input#service-name', serviceName);

    // 5. Next Step (Parameters)
    await page.click('button:has-text("Next")');

    // 6. Verify Schema Form
    // Should see "Service Configuration" and "Connection URL"
    await expect(page.getByText('Service Configuration')).toBeVisible();
    await expect(page.getByLabel('Connection URL')).toBeVisible();
    // Default is empty for Postgres in registry
    await expect(page.getByLabel('Connection URL')).toHaveValue('');

    // 7. Update Parameter
    await page.fill('input[type="password"]', 'postgresql://user:pass@db:5432/testdb');

    // 8. Next Step (Webhooks - Skip)
    await page.click('button:has-text("Next")');
    await expect(page.getByText('3. Webhooks & Transformers')).toBeVisible(); // Wait for Step 3

    // 9. Next Step (Auth - Skip)
    await page.click('button:has-text("Next")');
    await expect(page.getByText('4. Authentication')).toBeVisible(); // Wait for Step 4

    await page.click('button:has-text("Next")');

    // 10. Review & Finish
    await expect(page.getByText('5. Review & Finish')).toBeVisible();
    // Verify JSON preview contains the new URL (in plain text or masked?)
    // The review step usually shows JSON.
    // Let's just click Create.
    await page.click('button:has-text("Finish & Save to Local Marketplace")');

    // 11. Verify Toast
    await expect(page.getByText('Config Saved')).toBeVisible();

    // 12. Verify Toast (Mocked Backend)
    await expect(page.getByText('Config Saved').first()).toBeVisible();
  });
});
