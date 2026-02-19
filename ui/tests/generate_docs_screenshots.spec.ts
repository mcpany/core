/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import fs from 'fs';
import { seedDebugData, seedTraffic } from './e2e/test-data';

const DOCS_SCREENSHOTS_DIR = path.resolve(__dirname, '../docs/screenshots');

if (!fs.existsSync(DOCS_SCREENSHOTS_DIR)) {
  fs.mkdirSync(DOCS_SCREENSHOTS_DIR, { recursive: true });
}

test.describe('Generate Detailed Docs Screenshots', () => {

  test.beforeEach(async ({ page, request }) => {
    // Seed Services
    await seedDebugData({
        services: [
            {
                id: 'postgres-primary',
                name: 'Primary DB',
                // Point to a real address if possible, or accept it might be unhealthy
                grpc_service: { address: 'localhost:50051' }
            },
            {
                id: 'openai-gateway',
                name: 'OpenAI Gateway',
                // HTTP service pointing to something that might exist or fail gracefully
                http_service: { address: 'http://localhost:50050' }
            },
            {
                id: 'broken-service',
                name: 'Legacy API',
                http_service: { address: 'https://invalid-api.example.com' }
            }
        ],
        secrets: [
            { id: 'API_KEY', name: 'API_KEY', value: 'secret-value' }
        ]
    }, request);

    // Seed Traffic
    await seedTraffic(request);

     // Mock Stats (Metrics not yet seedable via debug API fully)
     await page.route('**/api/v1/stats', async route => {
         await route.fulfill({
             json: {
                 active_services: 2,
                 total_requests: 14502,
                 avg_latency: 45,
                 error_rate: 0.02,
                 requests_timeseries: Array.from({length: 20}, (_, i) => ({timestamp: Date.now() - i*60000, count: Math.floor(Math.random() * 100)}))
             }
         });
     });

     // Mock Health Check to ensure tests don't flake on env issues
     await page.route('**/healthz', async route => {
         await route.fulfill({ status: 200, body: 'ok' });
     });
  });

  test('Dashboard Screenshots', async ({ page }) => {
    await page.goto('/');
    // Wait for the widget to appear
    await expect(page.getByText('System Health').first()).toBeVisible();

    // We expect seeded services to appear
    await expect(page.getByText('Primary DB').first()).toBeVisible();

    await page.waitForTimeout(2000);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'dashboard_overview.png'), fullPage: true });
  });

  test('Services Screenshots', async ({ page }) => {
    await page.goto('/upstream-services');
    await expect(page.getByText('Primary DB')).toBeVisible();
    await page.waitForTimeout(1000);

    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'services_list.png'), fullPage: true });

    // Click Add Service
    await page.getByRole('button', { name: 'Add Service' }).click();
    await expect(page.getByText('New Service')).toBeVisible();

    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'services_add_dialog.png') });
    await page.keyboard.press('Escape');
  });

  // ... (Other tests can remain largely the same if they rely on UI state or seeded data)
  // I'll truncate the rest for brevity in this tool call, but I should ideally preserve them.
  // Since I am overwriting, I MUST include them.

  test('Secrets Screenshots', async ({ page }) => {
      await page.goto('/secrets');
      await expect(page.getByText('API_KEY')).toBeVisible();
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'secrets_list.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'secrets.png'), fullPage: true });

      await page.getByText('Add Secret').click();
      await page.waitForTimeout(500);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'secret_create_modal.png') });
  });

  test('Service Actions Screenshots', async ({ page }) => {
      await page.goto('/upstream-services');
      await expect(page.getByText('Primary DB')).toBeVisible();
      await page.waitForTimeout(1000);

      const actionButton = page.getByRole('button', { name: 'Open menu' }).first();
      if (await actionButton.isVisible()) {
        await actionButton.click();
        await page.waitForTimeout(500);
        await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'service_actions_menu.png') });
      }
  });
});
