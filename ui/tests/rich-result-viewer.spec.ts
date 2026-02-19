/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('Tool Inspector renders rich table result for complex data', async ({ page }) => {
  await page.goto('/tools');

  // Search for the test tool
  await page.getByPlaceholder('Search tools...').fill('get_complex_data');
  await expect(page.getByText('rich-result-test-service.get_complex_data').first()).toBeVisible();

  // Open inspector
  await page.getByRole('row', { name: 'rich-result-test-service.get_complex_data' }).getByRole('button', { name: 'Inspect' }).click();

  // Wait for inspector to open
  await expect(page.getByRole('dialog')).toBeVisible();
  await expect(page.getByRole('heading', { name: 'rich-result-test-service.get_complex_data' })).toBeVisible();

  // Execute tool (default args should work as they are empty object in seeded tool)
  await page.getByRole('button', { name: 'Execute' }).click();

  // Wait for result
  // Use precise selector to avoid matching service name "rich-result-test-service"
  await expect(page.locator('label').filter({ hasText: 'Result' })).toBeVisible();

  // Check if Table tab is active or available
  const tableTab = page.getByRole('tab', { name: 'Table' });
  await expect(tableTab).toBeVisible();

  // It might default to Table view because it's eligible
  // Verify content in table
  const table = page.getByRole('table');
  await expect(table).toBeVisible();

  // Verify data
  await expect(table.getByText('Alice')).toBeVisible();
  await expect(table.getByText('Bob')).toBeVisible();
  await expect(table.getByText('Admin')).toBeVisible();

  // Switch to JSON tab
  // Note: There might be multiple "JSON" tabs (one for schema, one for args, one for result)
  // We want the one in the result viewer. Since it's likely the last one rendered or scoped.
  // The tabs in RichResultViewer are: Table, JSON, Raw Output.
  // We can scope by finding the container.
  // Or just click the one that follows "Result".

  // Scoping to the result area
  const resultArea = page.locator('.grid', { hasText: 'Result' }).last();
  // Actually "Result" label is inside a grid div.

  // Let's try finding the tab list containing "Raw Output" which is unique to RichResultViewer
  const viewerTabs = page.locator('[role="tablist"]', { hasText: 'Raw Output' });
  await viewerTabs.getByRole('tab', { name: 'JSON' }).click();

  // Check for JSON content - look for specific value in pre/code
  await expect(page.getByText('"name": "Alice"')).toBeVisible();

  // Switch to Raw Output tab
  await viewerTabs.getByRole('tab', { name: 'Raw Output' }).click();
  await expect(page.getByText('"stdout":')).toBeVisible();
});
