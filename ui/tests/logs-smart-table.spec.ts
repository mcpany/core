/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Logs Smart Table View', () => {
  test.beforeAll(async ({ request }) => {
    // Seed realistic trace logs containing JSON arrays
    // This utilizes the backend seeder to write directly to the backend database
    // to test the "Broken Window" smart table fix without frontend mocking.
    const seedResponse = await request.post('/api/v1/debug/traces');
    expect(seedResponse.ok()).toBeTruthy();
  });

  test('should format JSON array log payload as a Smart Table instead of raw dump', async ({ page }) => {
    // Navigate directly to the global logs page since the seeded data isn't
    // bound to a specific upstream service created in the test context.
    await page.goto('/logs');

    // Wait for the logs to load into the view
    await expect(page.getByRole('heading', { name: 'Log Stream' })).toBeVisible({ timeout: 10000 });

    // The database seeder creates a trace for "search-tool" which outputs a JSON result with an array of strings.
    // Wait for the specific tool log we seeded, "search-tool"
    const row = page.locator('.group:has-text("search-tool")').first();
    await expect(row).toBeVisible({ timeout: 15000 });

    // Look for a row that has JSON content and click to expand it
    const jsonToggle = row.locator('button[aria-label="Expand JSON"]').first();
    await jsonToggle.waitFor({ state: 'visible', timeout: 5000 });
    await jsonToggle.click();

    // Look for a specific json structure
    // Since "search-tool" generates {"results": ["report_q3.pdf", "data_q3.xlsx"]}
    // We should see a Smart Table displaying these keys

    // Wait for the JSON viewer to expand
    // A smart table contains table headers
    const smartTable = row.locator('table').first();
    await expect(smartTable).toBeVisible({ timeout: 5000 });

    // The table should have "0" index and "value" columns since it's an array
    // inside "results", or if it's an object with array inside, SmartTable flattens it.
  });
});
