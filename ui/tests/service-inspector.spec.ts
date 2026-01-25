/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Service Inspector', () => {
  test('Service Inspector tab shows traces', async ({ page }) => {
    // Mock global endpoints
    await page.route('**/api/v1/doctor', async route => route.fulfill({ json: { status: 'healthy', checks: {} } }));
    await page.route('**/api/v1/services', async route => route.fulfill({ json: { services: [] } }));
    await page.route('**/api/v1/topology', async route => route.fulfill({ json: { nodes: [], edges: [] } }));
    await page.route('**/api/v1/alerts', async route => route.fulfill({ json: { alerts: [] } }));
    await page.route('**/api/v1/settings', async route => route.fulfill({ json: {} }));
    await page.route('**/api/v1/dashboard/top-tools', async route => route.fulfill({ json: [] }));
    await page.route('**/api/v1/dashboard/traffic', async route => route.fulfill({ json: [] }));

    // Mock gRPC to fail immediately so client falls back to REST
    await page.route('**/mcpany.api.v1.RegistrationService/GetService', async route => {
       await route.abort();
    });

    // Mock Service Detail with tools inside http_service as expected by client mapper
    await page.route('**/api/v1/services/test-service*', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          service: {
            id: 'test-service',
            name: 'Test Service',
            version: '1.0.0',
            http_service: {
                address: 'http://localhost:8080',
                tools: [{ name: 'test_tool' }]
            },
            disable: false
          }
        })
      });
    });

    // Mock Traces
    await page.route('**/api/traces', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          {
            id: 'trace-1',
            timestamp: new Date().toISOString(),
            totalDuration: 100,
            status: 'success',
            trigger: 'user',
            rootSpan: {
              id: 'span-1',
              name: 'tools/call',
              type: 'tool',
              startTime: Date.now(),
              endTime: Date.now() + 100,
              status: 'success',
              input: {
                method: 'tools/call',
                params: {
                  name: 'test_tool',
                  arguments: { foo: 'bar' }
                }
              },
              output: { result: 'ok' }
            }
          }
        ])
      });
    });

    // Mock Service Status
    await page.route('**/api/v1/services/test-service/status', async route => {
        await route.fulfill({ json: { metrics: {} } });
    });

    // Navigate
    await page.goto('/service/test-service');

    // Wait for header text
    await expect(page.locator('text=Test Service').first()).toBeVisible({ timeout: 15000 });

    // Click Inspector
    await page.getByRole('tab', { name: 'Inspector' }).click();

    // Verify "No traces found." disappears.
    // Use a loop to retry if it fails initially due to polling lag
    await expect(async () => {
        await expect(page.locator('text=No traces found.')).not.toBeVisible();
        await expect(page.locator('text=tools/call').first()).toBeVisible();
    }).toPass({ timeout: 15000 });

    // Click trace
    await page.locator('text=tools/call').first().click();

    // Verify details
    await expect(page.getByText('test_tool')).toBeVisible();
  });
});
