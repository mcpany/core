/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Audit Log Viewer', () => {
  test.beforeEach(async ({ page, request }) => {
    // Seed global state (like tools, services, and trigger discovery)
    await seedGlobalState(request);

    // Call the /api/v1/debug/traces/seed endpoint to populate the DB with a rich trace
    const seedRes = await request.post('/api/v1/debug/traces/seed');

    if (!seedRes.ok()) {
      const text = await seedRes.text();
      console.log(`DEBUG SEED FAILED: ${seedRes.status()} ${text}`);
    }

    if (!seedRes.ok()) {
      const text = await seedRes.text();
      console.log(`DEBUG SEED FAILED: ${seedRes.status()} ${text}`);
    }
    expect(seedRes.ok()).toBeTruthy();



    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', 'e2e-admin-core');
    await page.fill('input[name="password"]', 'password');
    await Promise.all([
      page.waitForURL('/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
    ]);
    await expect(page).toHaveURL('/', { timeout: 15000 });
  });

  test('should view seeded audit log with rich formatting', async ({ page }) => {
    // Navigate to Audit Logs page
    await page.goto('/audit');
    await expect(page).toHaveURL(/\/audit/);

    // Click "Filter" (Search) to load the most recent seeded log
    const filterBtn = page.getByRole('button', { name: 'Filter' });
    await filterBtn.click();

    // The seeded trace uses "orchestrator-task", wait for it
    await expect(page.getByText('orchestrator-task').first()).toBeVisible();

    // Click "View" on the first item
    const viewButtons = page.getByRole('button', { name: 'View' });
    await viewButtons.first().click();

    // Wait for the modal
    await expect(page.getByText('Audit Log Detail')).toBeVisible();
    await expect(page.getByText('Arguments')).toBeVisible();
    await expect(page.getByText('Result')).toBeVisible();

    // Verify JsonView for Arguments contains seeded data
    await expect(page.getByText('Analyze Q3 financial report')).toBeVisible();

    // Verify RichResultViewer for Result displays Table or Rendered data
    // Seeded data has a 'text' block (markdown) and tabular data. We can look for the "Table" and "Rendered" tabs
    await expect(page.getByRole('tab', { name: /Rendered/ })).toBeVisible();
    await expect(page.getByRole('tab', { name: /Table/ })).toBeVisible();

    // Tabular data has "month" "revenue" "target" headers
    await page.getByRole('tab', { name: /Table/ }).click();
    await expect(page.getByText('revenue').first()).toBeVisible();

    // Rendered data has markdown text
    await page.getByRole('tab', { name: /Rendered/ }).click();
    await expect(page.getByText('Q3 Financial Report').first()).toBeVisible();
  });
});
