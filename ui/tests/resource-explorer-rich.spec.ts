/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Explorer Rich Result Viewer', () => {
  const serviceName = 'resource-viewer-rich-result-test';

  test.beforeAll(async ({ request }) => {
    // Clean up
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => { });

    // Seed service
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        command_line_service: {
          command: 'echo',
          resources: [
            { uri: 'test://data.json', name: 'JSON Data', mimeType: 'application/json' },
            { uri: 'test://invalid.json', name: 'Invalid JSON', mimeType: 'application/json' }
          ],
          reads: {
            'test://data.json': {
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
            },
            'test://invalid.json': {
              contents: [
                {
                  uri: 'test://invalid.json',
                  mimeType: 'application/json',
                  text: '{ invalid json '
                }
              ]
            }
          }
        }
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => { });
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
