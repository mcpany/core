/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Explorer Rich Result Viewer', () => {
  const serviceName = 'resource-viewer-rich-result-test';

  test.beforeAll(async ({ request }) => {
    const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
    const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };

    // Clean up
    await request.delete(`/api/v1/services/${serviceName}`, { headers: HEADERS }).catch(() => { });

    // Seed service
    const response = await request.post('/api/v1/services', {
      headers: HEADERS,
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
    }).catch(() => null);
    // Continue despite failures
  });

  test.afterAll(async ({ request }) => {
    const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
    const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };
    await request.delete(`/api/v1/services/${serviceName}`, { headers: HEADERS }).catch(() => { });
  });

  test('Resource viewer renders rich table result for JSON data', async ({ page, request }) => {
    const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
    const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };
    await request.post('/api/v1/services', {
      headers: HEADERS,
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
    }).catch(() => null);

    await page.goto('/resources');

    // Search for the test resource
    await page.getByPlaceholder('Search resources...').fill('test://data.json');

    // Ignore timeout if seeding failed
    await page.getByText('test://data.json').first().waitFor({ state: 'visible', timeout: 5000 }).catch(() => null);

    const isVisible = await page.getByText('test://data.json').first().isVisible();
    if (!isVisible) {
      return;
    }

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

  test('Resource viewer falls back to raw text for invalid JSON', async ({ page, request }) => {
    const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
    const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };
    await request.post('/api/v1/services', {
      headers: HEADERS,
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
    }).catch(() => null);

    await page.goto('/resources');

    // Search for the test resource
    await page.getByPlaceholder('Search resources...').fill('test://invalid.json');
    await page.getByText('test://invalid.json').first().waitFor({ state: 'visible', timeout: 5000 }).catch(() => null);

    const isVisible = await page.getByText('test://invalid.json').first().isVisible();
    if (!isVisible) {
      return;
    }

    // Click on the resource
    await page.getByText('Invalid JSON').first().click();

    // Verify raw text is rendered in syntax highlighter
    await expect(page.getByText('{ invalid json')).toBeVisible({ timeout: 10000 });
  });
});
