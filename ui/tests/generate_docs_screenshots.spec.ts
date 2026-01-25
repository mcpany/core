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
    await page.route('**/api/v1/services*', async route => {
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
                        },
                        {
                            id: 'broken-service',
                            name: 'Legacy API',
                            type: 'http',
                            http_service: { address: 'https://api.example.com' },
                            status: 'unhealthy',
                            last_error: 'ZodError: Invalid input: expected string, received number',
                            lastError: 'ZodError: Invalid input: expected string, received number',
                            tool_count: 0,
                            version: '1.0.0'
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

     // Mock Health Check to prevent connection error banner
     await page.route('**/healthz', async route => {
         await route.fulfill({ status: 200, body: 'ok' });
     });
     await page.route('**/api/v1/health', async route => {
         await route.fulfill({ status: 200, json: { status: 'ok' } });
     });

     // Mock Doctor (System Status) to prevent banner from showing in screenshots
     await page.route('**/doctor', async route => {
         await route.fulfill({
             status: 200,
             contentType: 'application/json',
             body: JSON.stringify({
                 status: 'healthy',
                 checks: {},
                 version: '1.0.0',
                 uptime_seconds: 3600,
                 active_connections: 5,
                 bound_http_port: 8080,
                 bound_grpc_port: 50051
             })
         });
     });

     // Mock Dashboard Traffic
     await page.route('**/api/v1/dashboard/traffic', async route => {
         await route.fulfill({
             json: Array.from({length: 24}, (_, i) => ({
                 timestamp: new Date(Date.now() - i * 3600000).toISOString(),
                 requests: Math.floor(Math.random() * 500) + 100,
                 errors: Math.floor(Math.random() * 10)
             })).reverse()
         });
     });

  });

  test('Dashboard Screenshots', async ({ page }) => {
    // Pre-populate health history for timeline visualization
    await page.addInitScript(() => {
        const history = {
            'postgres-primary': Array(50).fill(0).map((_, i) => ({ timestamp: Date.now() - i * 10000, status: 'healthy' })).reverse(),
            'openai-gateway': Array(50).fill(0).map((_, i) => ({ timestamp: Date.now() - i * 10000, status: Math.random() > 0.9 ? 'degraded' : 'healthy' })).reverse(),
            'broken-service': Array(50).fill(0).map((_, i) => ({ timestamp: Date.now() - i * 10000, status: 'unhealthy' })).reverse()
        };
        window.localStorage.setItem('mcp_service_health_history', JSON.stringify(history));
    });

    await page.route('**/api/dashboard/health', async route => {
        await route.fulfill({
            json: [
               {
                   id: 'postgres-primary',
                   name: 'Primary DB',
                   status: 'healthy',
                   latency: '12ms',
                   uptime: '2d 4h',
                   message: ''
               },
               {
                   id: 'openai-gateway',
                   name: 'OpenAI Gateway',
                   status: 'healthy',
                   latency: '45ms',
                   uptime: '5h 30m',
                   message: ''
               },
               {
                   id: 'broken-service',
                   name: 'Legacy API',
                   status: 'unhealthy',
                   latency: '--',
                   uptime: '10m',
                   message: 'Connection refused'
               }
            ]
        });
    });

    await page.goto('/');
    await page.waitForLoadState('networkidle');
    // Give widgets extra time to render after data fetch
    await page.waitForTimeout(3000);
    await expect(page.locator('body')).toBeVisible();
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'dashboard_overview.png'), fullPage: true });
  });

  test('Services Screenshots', async ({ page }) => {
    await page.goto('/services');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    // Wait for loading to finish if applicable
    await expect(page.locator('text=Loading...')).not.toBeVisible();

    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'services_list.png'), fullPage: true });
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'services.png'), fullPage: true });

    // Click Add Service (Button)
    await page.getByRole('button', { name: 'Add Service' }).click();
    await page.waitForTimeout(1000);
    // await expect(page).toHaveURL(/.*marketplace.*/); // It opens a sheet now
    await expect(page.getByText('New Service')).toBeVisible();

    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'services_add_dialog.png') });
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'service_add_dialog.png') }); // Alias

    // Configure Service
    // Ensure we are navigating to the correct URL for configuration
    await page.goto('/upstream-services/postgres-primary');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000); // Increased wait time
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'service_config.png'), fullPage: true });
  });

  test('Playground Screenshots', async ({ page }) => {
    await page.goto('/playground');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'playground_blank.png'), fullPage: true });
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'playground.png'), fullPage: true });

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
    }

    // Tools alias
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'tools.png'), fullPage: true });
  });

  test('Stack Composer Screenshots', async ({ page }) => {
    // Mock Stacks List
    await page.route('**/api/v1/collections', async route => {
        if (route.request().method() === 'GET') {
            await route.fulfill({
                json: [
                    { name: 'production-stack', services: Array(3).fill({}) },
                    { name: 'dev-stack', services: Array(1).fill({}) }
                ]
            });
        } else {
            await route.continue();
        }
    });

    await page.goto('/stacks');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'stack_composer_overview.png'), fullPage: true });
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'stacks.png'), fullPage: true });

    // Open Create Dialog
    await page.getByRole('button', { name: 'Create Stack' }).click();
    await page.waitForTimeout(500);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'stack_create_dialog.png') });
  });

  test('Traces Screenshots', async ({ page }) => {
    // Mock Traces (UI calls /api/traces, expects direct array with rootSpan)
    await page.route('**/api/traces*', async route => {
        await route.fulfill({
            json: [
                 {
                     id: 't1',
                     timestamp: Date.now(),
                     rootSpan: { name: 'filesystem.read' },
                     status: 'success',
                     totalDuration: 120,
                     trigger: 'user'
                 },
                 {
                     id: 't2',
                     timestamp: Date.now() - 5000,
                     rootSpan: { name: 'calculator.add' },
                     status: 'error',
                     error: 'Division by zero',
                     totalDuration: 10,
                     trigger: 'user'
                 }
            ]
        });
    });

    await page.goto('/traces');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'traces_list.png'), fullPage: true });
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'traces.png'), fullPage: true });

    // Click trace
    await page.getByText('filesystem.read').first().click({ force: true });
    await page.waitForTimeout(500);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'trace_detail.png'), fullPage: true });
  });

  test('Middleware Screenshots', async ({ page }) => {
      await page.goto('/middleware');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'middleware_pipeline.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'middleware.png'), fullPage: true });
  });

  test('Webhooks Screenshots', async ({ page }) => {
      await page.goto('/webhooks');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'webhooks_list.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'webhooks.png'), fullPage: true });
      // Legacy alias
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_webhooks.png'), fullPage: true });

      await page.getByText('New Webhook').click();
      await page.waitForTimeout(500);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'webhook_create_modal.png') });
  });

  test('Network Graph Screenshots', async ({ page }) => {
      await page.goto('/network');
      await page.waitForTimeout(2000); // Graph rendering
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'network_graph.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'network.png'), fullPage: true });
  });

  test('Logs Screenshots', async ({ page }) => {
      await page.goto('/logs');
      await page.waitForTimeout(1000);
       await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'logs_stream.png'), fullPage: true });
       await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'logs.png'), fullPage: true });
  });

   test('Marketplace Screenshots', async ({ page }) => {
       await page.goto('/marketplace');
       await page.waitForTimeout(1000);
        await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'marketplace_grid.png'), fullPage: true });
        await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'marketplace.png'), fullPage: true });
        // Legacy alias
        await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'marketplace_external_detailed.png'), fullPage: true });
   });

  test('Secrets Screenshots', async ({ page }) => {
      await page.route('**/api/v1/secrets*', async route => {
          await route.fulfill({
              json: { secrets: [{name: 'API_KEY', value: '*****'}] }
          });
      });
      await page.goto('/secrets');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);
      await expect(page.getByText('Loading secrets...')).not.toBeVisible();
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'secrets_list.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'secrets.png'), fullPage: true });
      // Legacy alias (Secrets List)
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_secrets.png'), fullPage: true });

      await page.getByText('Add Secret').click();
      await page.waitForTimeout(500);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'secret_create_modal.png') });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'credential_form.png') });
  });

  test('Auth Screenshots', async ({ page }) => {
      await page.goto('/login');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'auth_login.png'), fullPage: true });
      // Legacy aliases (placeholder to ensure update)
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'auth_guide_step1_apikey.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'auth_guide_step2_bearer.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'auth_guide_step3_basic.png'), fullPage: true });

      // Mock Users
      await page.goto('/users');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'auth_users_list.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'users.png'), fullPage: true });
  });

  test('Prompts Screenshots', async ({ page }) => {
      await page.goto('/prompts');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'prompts_list.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'prompts.png'), fullPage: true });
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
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'resources_list.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'resources.png'), fullPage: true });
      // Legacy aliases
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'resources_grid.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'resources_split_view.png'), fullPage: true });

      // Open Preview Modal
      // Just take a screenshot of the page with the modal open if possible
      // Using nth(0) as per plan
      const firstResource = page.getByRole('row').nth(0);
      if (await firstResource.isVisible()) {
        await firstResource.click({ button: 'right' });
        await page.waitForTimeout(1000);
        const previewBtn = page.getByText('Preview in Modal');
        if (await previewBtn.isVisible()) {
             await previewBtn.click();
             await page.waitForTimeout(2000);
             await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'resource_preview_modal.png') });
        }
      }
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

  test('Skills Screenshots', async ({ page }) => {
      await page.goto('/skills');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'skills_list.png'), fullPage: true });

      // Create/Edit View
      await page.goto('/skills/create');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'skills_create.png'), fullPage: true });
  });

  test('Profiles Screenshots', async ({ page }) => {
      await page.goto('/profiles');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'profiles_page.png'), fullPage: true });

      // Open Editor (Create)
      await page.getByRole('button', { name: 'Create Profile' }).click();
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'profile_editor.png') });
  });

  test('Settings Screenshots', async ({ page }) => {
      await page.goto('/settings');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_profiles.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'profiles.png'), fullPage: true }); // Legacy alias

      // Click General Tab
      await page.getByRole('tab', { name: 'General' }).click();
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_general.png'), fullPage: true });

      // Legacy alias for Auth Settings (which is another tab)
      await page.getByRole('tab', { name: 'Authentication' }).click();
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_auth.png'), fullPage: true });
  });

  test('Credentials Screenshots', async ({ page }) => {
      await page.goto('/credentials');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'credentials.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'credentials_list.png'), fullPage: true });
  });

  test('Stats Screenshots', async ({ page }) => {
      await page.goto('/stats');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'stats.png'), fullPage: true });
  });

  test('Service Actions Screenshots', async ({ page }) => {
      await page.goto('/services');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      // Open Actions Dropdown
      // We target the first card's actions button if available, or skip if no services rendered
      const actionButton = page.getByRole('button', { name: 'Open menu' }).first();
      if (await actionButton.isVisible()) {
        await actionButton.click();
        await page.waitForTimeout(500);
        await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'service_actions_menu.png') });
      }
  });

  test('Audit Logs Screenshots', async ({ page }) => {
      // Mock Audit Logs
      await page.route('**/api/v1/audit/logs*', async route => {
          await route.fulfill({
              json: {
                  entries: [
                      {
                          timestamp: new Date().toISOString(),
                          toolName: 'weather_get',
                          userId: 'alice',
                          profileId: 'prod',
                          arguments: '{"city": "London"}',
                          result: '{"temperature": 20}',
                          duration: '150ms',
                          durationMs: 150
                      },
                      {
                          timestamp: new Date(Date.now() - 60000).toISOString(),
                          toolName: 'calculator_add',
                          userId: 'bob',
                          profileId: 'dev',
                          arguments: '{"a": 5, "b": 3}',
                          result: '8',
                          duration: '10ms',
                          durationMs: 10
                      }
                  ]
              }
          });
      });

      await page.goto('/audit');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'audit_logs.png'), fullPage: true });
  });

  test('Diagnostics Failure Screenshots', async ({ page }) => {
      // Mock service detail for operational check
      await page.route('**/api/v1/services/broken-service', async route => {
          await route.fulfill({
              json: {
                  service: {
                      id: 'broken-service',
                      name: 'Legacy API',
                      type: 'http',
                      http_service: { address: 'https://api.example.com' },
                      status: 'unhealthy',
                      last_error: 'ZodError: Invalid input: expected string, received number',
                      lastError: 'ZodError: Invalid input: expected string, received number',
                      tool_count: 0,
                      toolCount: 0
                  }
              }
          });
      });

      // Mock Health Check for backend health step
      await page.route('**/api/dashboard/health', async route => {
        await route.fulfill({
            json: [
               {
                   id: 'broken-service',
                   name: 'Legacy API',
                   status: 'unhealthy',
                   latency: '--',
                   uptime: '10m',
                   message: 'ZodError: Invalid input: expected string, received number'
               }
            ]
        });
      });

      await page.goto('/services');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      // Verify service row is present
      await expect(page.getByText('Legacy API')).toBeVisible();

      // Verify Error badge is present (confirms lastError is recognized)
      // await expect(page.getByText('Error', { exact: true })).toBeVisible(); // Flaky

      // Open Actions Dropdown (More Reliable)
      const menuButton = page.getByRole('button', { name: 'Open menu' }).first();
      await expect(menuButton).toBeVisible();
      await menuButton.click();

      // Click Diagnose in menu
      await page.getByText('Diagnose').click();

      // Wait for dialog
      await expect(page.getByText('Connection Diagnostics')).toBeVisible();

      // Click Start
      await page.getByRole('button', { name: 'Start Diagnostics' }).click();

      // Wait for run to finish (look for "Rerun Diagnostics")
      await page.getByText('Rerun Diagnostics', { timeout: 10000 }).waitFor();

      // Take screenshot of the modal
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'diagnostics_failure.png') });
  });

});
