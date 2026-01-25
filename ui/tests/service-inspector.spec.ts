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

    // Mock Service Detail
    await page.route('**/api/v1/services/test-service', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          service: {
            id: 'test-service',
            name: 'Test Service',
            version: '1.0.0',
            tools: [{ name: 'test_tool' }],
            disable: false,
            httpService: { address: 'http://localhost:8080' }
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

    // Verify trace list item
    await expect(page.locator('button', { hasText: 'tools/call' })).toBeVisible();

    // Click trace
    await page.locator('button', { hasText: 'tools/call' }).click();

    // Verify details
    await expect(page.getByText('test_tool')).toBeVisible();
  });
});
