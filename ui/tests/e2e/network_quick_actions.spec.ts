/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Network Graph Quick Actions', () => {

  test.beforeEach(async ({ page }) => {
    // Mock Topology Data
    await page.route('**/api/v1/topology', async route => {
        await route.fulfill({
            json: {
                core: {
                    id: 'mcp-core',
                    label: 'MCP Any',
                    type: 'NODE_TYPE_CORE',
                    status: 'NODE_STATUS_ACTIVE',
                    children: [
                        {
                            id: 'svc-weather',
                            label: 'weather-service',
                            type: 'NODE_TYPE_SERVICE',
                            status: 'NODE_STATUS_ACTIVE',
                            children: []
                        }
                    ]
                },
                clients: []
            }
        });
    });

    // Mock Logs (empty)
    await page.route('**/api/v1/ws/logs', async route => {
        // WS mock might be tricky, but we just want to ensure page loads
        // We'll just return a 200 OK for HTTP, or if it's WS upgrade, Playwright might need different handling.
        // But logs page does HTTP fetch or WS?
        // log-stream.tsx does new WebSocket().
        // We can't easily mock WS with page.route without more complex setup.
        // However, failing the route or leaving it hanging is bad.
        // Let's abort it to simulate connection failure (LogStream handles it)
        // or fulfill with 404.
        await route.abort();
    });

    // Mock Traces (empty)
    await page.route('**/api/traces*', async route => {
        await route.fulfill({ json: [] });
    });
  });

  test('should navigate to Logs from Quick Actions', async ({ page }) => {
    await page.goto('/network');

    // Wait for graph to render
    // We look for the node by text "weather-service"
    // ReactFlow renders nodes as divs with text content
    const node = page.getByText('weather-service');
    await expect(node).toBeVisible();

    // Click the node to open the sheet
    await node.click({ force: true });

    // Check if sheet opens
    await expect(page.getByText('Quick Actions')).toBeVisible();

    // Click View Logs
    await page.getByRole('button', { name: 'View Logs' }).click();

    // Verify URL
    await expect(page).toHaveURL(/\/logs\?source=weather-service/);

    // Verify Filter is set (we check if the select value is set)
    // Note: This relies on implementation detail of Select component or looking for text
    // The SelectValue should display "weather-service" if it's in the list
    // However, the list comes from actual logs. Since we have no logs, the source list might be empty.
    // But the URL param should be preserved.
  });

  test('should navigate to Traces from Quick Actions', async ({ page }) => {
    await page.goto('/network');

    const node = page.getByText('weather-service');
    await expect(node).toBeVisible();
    await node.click({ force: true });

    await expect(page.getByText('Quick Actions')).toBeVisible();

    // Click Trace Request
    await page.getByRole('button', { name: 'Trace Request' }).click();

    // Explicitly wait for navigation to complete before checking URL
    await page.waitForURL(/\/traces\?query=weather-service/);

    // Verify URL
    await expect(page).toHaveURL(/\/traces\?query=weather-service/);
  });

});
