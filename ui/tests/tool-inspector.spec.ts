/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser, cleanupUser } from './e2e/test-data';

test('Tools page loads and inspector opens', async ({ page, request }) => {
  await seedServices(request);
  await seedUser(request, "e2e-tool-admin");
  await page.goto('/login');
  await page.fill('input[name="username"]', "e2e-tool-admin");
  await page.fill('input[name="password"]', 'password');
  await Promise.all([
      page.waitForURL('/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
  ]);

  await page.goto('/tools');

  // Check if tools are listed
  await expect(page.getByText('calculator')).toBeVisible();

  // Open inspector for calculator
  await page.locator('tr').filter({ hasText: 'calculator' }).getByText('Inspect').click();

  // Check if inspector sheet is open (Wait for title)
  await expect(page.getByText('calculator').first()).toBeVisible();

  // Switch to Schema tab
  await page.getByRole('tab', { name: 'Schema' }).click();

  // Switch to JSON sub-tab to verify raw schema
  // Ensure the Schema tab content is visible first
  const schemaPanel = page.getByRole('tabpanel').filter({ hasText: 'Visual' });
  await expect(schemaPanel).toBeVisible();

  // Click the JSON trigger inside the schema content
  await schemaPanel.getByRole('tab', { name: 'JSON' }).click();

  // The schema content from seed: { type: "object", properties: { a: { type: "number" }, b: { type: "number" } } }
  // We check for "a" property in the JSON view
  await expect(page.locator('pre').filter({ hasText: /"a"/ })).toBeVisible();
  await expect(page.locator('pre').filter({ hasText: /"type": "object"/ })).toBeVisible();

  // Verify service name is shown in details (Scoped to the sheet)
  await expect(page.locator('div[role="dialog"]').getByText('svc_03')).toBeVisible();

  await cleanupServices(request);
  await cleanupUser(request, "e2e-tool-admin");
});
