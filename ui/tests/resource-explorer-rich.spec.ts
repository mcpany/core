/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Explorer Rich Result Viewer', () => {
  const serviceName = 'resource-viewer-rich-result-test';

  test.beforeEach(async ({ page }) => {
    // Mock the resources endpoint
    await page.route('**/api/v1/resources', async (route) => {
      await route.fulfill({
        json: {
          resources: [
            { uri: 'test://data.json', name: 'JSON Data', mimeType: 'application/json', serviceId: serviceName },
            { uri: 'test://invalid.json', name: 'Invalid JSON', mimeType: 'application/json', serviceId: serviceName }
          ]
        }
      });
    });

    // Mock resource read
    await page.route('**/api/v1/resources/read*', async (route) => {
      const urlObj = new URL(route.request().url());
      const uri = urlObj.searchParams.get('uri');

      if (uri === 'test://data.json') {
        await route.fulfill({
          json: {
            contents: [
              {
                uri: 'test://data.json',
                mimeType: 'application/json',
                text: JSON.stringify([
                  { name: 'Alice', role: 'Admin', id: 1 },
                  { name: 'Bob', role: 'User', id: 2 }
                ])
              }
            ]
          }
        });
      } else if (uri === 'test://invalid.json') {
        await route.fulfill({
          json: {
            contents: [
              {
                uri: 'test://invalid.json',
                mimeType: 'application/json',
                text: '{ invalid json '
              }
            ]
          }
        });
      } else {
        await route.fulfill({ status: 404 });
      }
    });
  });

  test('Resource viewer renders rich table result for JSON data', async ({ page }) => {
    await page.goto('/resources');

    // Search for the test resource
    await page.getByPlaceholder('Search resources...').fill('test://data.json');
    await expect(page.getByText('test://data.json').first()).toBeVisible({ timeout: 10000 });

    // Click on the resource
    await page.getByText('JSON Data').first().click();

    // Verify RichResultViewer is rendered
    // Wait for the Table tab to be visible
    const tableTab = page.getByRole('tab', { name: 'Table' });
    await expect(tableTab).toBeVisible({ timeout: 10000 });

    // Verify content in table
    const table = page.getByRole('table');
    await expect(table).toBeVisible();

    // Verify data
    await expect(table.getByText('Alice')).toBeVisible();
    await expect(table.getByText('Bob')).toBeVisible();
    await expect(table.getByText('Admin')).toBeVisible();
  });

  test('Resource viewer falls back to raw text for invalid JSON', async ({ page }) => {
    await page.goto('/resources');

    // Search for the test resource
    await page.getByPlaceholder('Search resources...').fill('test://invalid.json');
    await expect(page.getByText('test://invalid.json').first()).toBeVisible({ timeout: 10000 });

    // Click on the resource
    await page.getByText('Invalid JSON').first().click();

    // Verify raw text is rendered in syntax highlighter
    await expect(page.getByText('{ invalid json')).toBeVisible({ timeout: 10000 });
  });
});
