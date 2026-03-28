/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Explorer Rich Result Viewer', () => {
  const serviceName = 'resource-viewer-rich-result-test';

  test.beforeAll(async ({ request }) => {
    // Clean up
    await request.delete("/api/v1/services/" + serviceName).catch(() => { });

    // Seed service
    const response = await request.post('/api/v1/services', {
      data: {
        id: serviceName,
        name: serviceName,
        command_line_service: {
          command: 'echo',
          args: ['hello'],
          resources: [
            {
              uri: 'test://data.json',
              name: 'JSON Data',
              mime_type: 'application/json',
              static: {
                text_content: JSON.stringify([
                  { name: 'Alice', role: 'Admin', id: 1 },
                  { name: 'Bob', role: 'User', id: 2 }
                ])
              }
            },
            {
              uri: 'test://invalid.json',
              name: 'Invalid JSON',
              mime_type: 'application/json',
              static: {
                text_content: '{ invalid json '
              }
            }
          ]
        }
      }
    });
    if (!response.ok()) {
        console.error('SEED FAILED', await response.text());
    }
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    await request.delete("/api/v1/services/" + serviceName).catch(() => { });
  });

  test('Resource viewer renders rich table result for JSON data', async ({ page }) => {
    await page.goto('/resources');

    // Wait for the list to load
    const resourceItem = page.getByText('JSON Data').first();
    await expect(resourceItem).toBeVisible({ timeout: 30000 });
    await resourceItem.click();

    // Verify content in table
    const table = page.getByRole('table');
    await expect(table.getByText('Alice')).toBeVisible({ timeout: 15000 });
    await expect(table.getByText('Bob')).toBeVisible();
  });

  test('Resource viewer falls back to raw text for invalid JSON', async ({ page }) => {
    await page.goto('/resources');

    // Wait for the list to load
    const resourceItem = page.getByText('Invalid JSON').first();
    await expect(resourceItem).toBeVisible({ timeout: 30000 });
    await resourceItem.click();

    // Verify raw text is rendered
    await expect(page.getByText('{ invalid json')).toBeVisible({ timeout: 15000 });
  });
});
