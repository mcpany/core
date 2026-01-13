/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import fs from 'fs';

const DOCS_SCREENSHOTS_DIR = path.resolve(__dirname, '../docs/screenshots');

if (!fs.existsSync(DOCS_SCREENSHOTS_DIR)) {
  fs.mkdirSync(DOCS_SCREENSHOTS_DIR, { recursive: true });
}

test.describe('Generate Detailed Docs Screenshots', () => {

  test.beforeEach(async ({ page }) => {
    // Global mocks to ensure consistent state
    await page.route('**/api/v1/services', async route => {
        if (route.request().method() === 'GET') {
            await route.fulfill({
                json: {
                    services: [
                        {
                            id: 'postgres-primary',
                            name: 'Primary DB',
                            type: 'remote',
                            endpoint: 'grpc://postgres:5432',
                            status: 'healthy',
                            uptime: '2d 4h',
                            version: '1.0.0'
                        },
                        {
                            id: 'openai-gateway',
                            name: 'OpenAI Gateway',
                            type: 'mcp',
                            endpoint: 'http://openai-mcp:8080',
                            status: 'healthy',
                            uptime: '5h 30m',
                            version: '2.1.0'
                        }
                    ]
                }
            });
        } else {
            await route.continue();
        }
    });

    await page.route('**/api/v1/services/postgres-primary', async route => {
        await route.fulfill({
            json: {
                service: {
                    id: 'postgres-primary',
                    name: 'Primary DB',
                    type: 'remote',
                    endpoint: 'grpc://postgres:5432',
                    status: 'healthy',
                    config: {
                         env: { 'DB_PASS': '********' }
                    }
                }
            }
        });
    });

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

     // Mock Logs
     await page.route('**/api/v1/logs/stream**', async route => {
         // This might be WS, but if HTTP fallback:
         await route.fulfill({ json: [] });
     });

  });

  test('Dashboard Screenshots', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    await expect(page.locator('body')).toBeVisible();
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'dashboard_overview.png'), fullPage: true });
  });

  test('Services Screenshots', async ({ page }) => {
    await page.goto('/services');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'services_list.png'), fullPage: true });

    // Click Add Service
    await page.getByRole('button', { name: 'Add Service' }).click();
    await page.waitForTimeout(500);
    await expect(page.getByText('New Service')).toBeVisible();
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'services_add_dialog.png') });

    // Close
    await page.keyboard.press('Escape');

    // Configure Service
    // We navigate directly
    await page.goto('/services/postgres-primary');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'service_config.png'), fullPage: true });
  });

  test('Playground Screenshots', async ({ page }) => {
    await page.goto('/playground');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'playground_blank.png'), fullPage: true });

    // Mock Tools
    await page.route('**/api/v1/tools', async route => {
         await route.fulfill({
             json: {
                 tools: [
                     { name: 'filesystem.list_dir', description: 'List files in directory', inputSchema: { type: 'object', properties: { path: { type: 'string' } }, required: ['path'] } },
                     { name: 'calculator.add', description: 'Add two numbers', inputSchema: { type: 'object', properties: { a: { type: 'number' }, b: { type: 'number' } }, required: ['a', 'b'] } }
                 ]
             }
         });
    });

    // Reload to get tools
    await page.reload();
    await page.waitForTimeout(1000);

    // Select Tool
    const tool = page.getByText('filesystem.list_dir');
    if (await tool.isVisible()) {
        await tool.click();
        await page.waitForTimeout(500);
        await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'playground_tool_selected.png'), fullPage: true });

        // Fill Form
        await page.getByLabel('path').fill('/var/log');
        await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'playground_form_filled.png'), fullPage: true });

        // Mock Result (Implicit via executed action simulation visuals if we could run it, but just form is enough for now or we fake result)
        // We can't easily fake the execution result without backend, but we can fake the UI state if we manual inject HTML? Too complex.
        // We will stick to form filled.
    }
  });

  test('Stack Composer Screenshots', async ({ page }) => {
    await page.goto('/stacks');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'stack_composer_overview.png'), fullPage: true });

    // Click New Stack (if exists) or just the editor view
    // Assuming /stacks goes to list, and we need /stacks/new or similar
    // If /stacks is the editor:
    // Check if we have 'Service Palette'
    if (await page.getByText('Service Palette').isVisible()) {
        await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'stack_composer_palette.png'), fullPage: true });
    }
  });

  test('Traces Screenshots', async ({ page }) => {
    // Mock Traces
    await page.route('**/api/v1/traces', async route => {
        await route.fulfill({
            json: {
                 traces: [
                     { id: 't1', timestamp: Date.now(), tool: 'filesystem.read', status: 'success', duration: 120 },
                     { id: 't2', timestamp: Date.now() - 5000, tool: 'calculator.add', status: 'error', error: 'Division by zero', duration: 10 }
                 ]
            }
        });
    });

    await page.goto('/traces');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'traces_list.png'), fullPage: true });

    // Click trace
    await page.getByText('filesystem.read').click();
    await page.waitForTimeout(500);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'trace_detail.png'), fullPage: true });
  });

  test('Middleware Screenshots', async ({ page }) => {
      await page.goto('/middleware');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'middleware_pipeline.png'), fullPage: true });
  });

  test('Webhooks Screenshots', async ({ page }) => {
      await page.goto('/webhooks');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'webhooks_list.png'), fullPage: true });

      await page.getByText('Add Webhook').click();
      await page.waitForTimeout(500);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'webhook_create_modal.png') });
  });

  test('Network Graph Screenshots', async ({ page }) => {
      await page.goto('/network');
      await page.waitForTimeout(2000); // Graph rendering
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'network_graph.png'), fullPage: true });
  });

  test('Logs Screenshots', async ({ page }) => {
      await page.goto('/logs');
      await page.waitForTimeout(1000);
       await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'logs_stream.png'), fullPage: true });
  });

   test('Marketplace Screenshots', async ({ page }) => {
      await page.goto('/marketplace');
      await page.waitForTimeout(1000);
       await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'marketplace_grid.png'), fullPage: true });
  });

  test('Secrets Screenshots', async ({ page }) => {
      await page.route('**/api/v1/secrets', async route => {
          await route.fulfill({
              json: { secrets: [{name: 'API_KEY', value: '*****'}] }
          });
      });
      await page.goto('/secrets');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'secrets_list.png'), fullPage: true });

      await page.getByText('Add Secret').click();
      await page.waitForTimeout(500);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'secret_create_modal.png') });
  });

  test('Auth Screenshots', async ({ page }) => {
      await page.goto('/login');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'auth_login.png'), fullPage: true });

      // Mock Users
      await page.goto('/users');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'auth_users_list.png'), fullPage: true });
  });

  test('Prompts Screenshots', async ({ page }) => {
      await page.goto('/prompts');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'prompts_list.png'), fullPage: true });
  });

  test('Search Screenshots', async ({ page }) => {
       await page.goto('/');
       await page.waitForTimeout(1000);
       await page.keyboard.press('Control+k');
       await page.waitForTimeout(500);
       await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'global_search.png') });
  });

  test('Resources Screenshots', async ({ page }) => {
      await page.goto('/resources');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'resources_list.png'), fullPage: true });

      // select one
      // If mock exists from previous run or we assume empty.
      // We need to ensuring mock exists if we rely on it.
  });

  test('Alerts Screenshots', async ({ page }) => {
      await page.goto('/alerts');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'alerts_list.png'), fullPage: true });
  });

  test('Mobile Screenshots', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 812 });
      await page.goto('/');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'mobile_dashboard.png'), fullPage: true });
  });


});
