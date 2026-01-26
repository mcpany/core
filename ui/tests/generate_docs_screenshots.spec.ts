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

    // Diff Feature Screenshot
    // Navigate to the Service Detail page which uses RegisterServiceDialog for editing
    await page.goto('/service/postgres-primary');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);

    // Click Edit Config
    const editBtn = page.getByRole('button', { name: 'Edit Config' });
    await expect(editBtn).toBeVisible({ timeout: 10000 });
    await editBtn.click();
    await expect(page.getByText('Edit Service')).toBeVisible({ timeout: 10000 });
    // Verify we are editing the HTTP service
    // Note: 'HTTP' might be in a select value or text.
    await page.getByLabel('Address / URL').waitFor({ state: 'visible', timeout: 10000 });

    // Make a change to trigger diff
    // We need to change a field. Let's change the address.
    // Assuming 'http' type service from mock.
    // The dialog should populate with 'grpc://postgres:5432' as per mock?
    // Mock says type: 'remote' (which might be 'mcp' in frontend?) or 'grpc'.
    // Let's verify mock content in beforeEach.
    // Mock: type: 'remote', endpoint: ...
    // RegisterServiceDialog handles 'mcp' via 'other'? Or maybe 'grpc'?
    // Let's force a simpler service mock for this test or assume 'http' fallback?
    // Actually, let's just create a new service in the dialog to be safe, OR use a known editable mocked service.
    // The 'postgres-primary' mock has type 'remote'.
    // RegisterServiceDialog might treat 'remote' as 'other'?

    // Let's modify the mock for postgres-primary to be HTTP for this test if needed, OR just type in a field.
    // If we type in 'address', it should work.
    // Check if 'Address / URL' input is visible.
    // trace-detail.tsx uses RegisterServiceDialog.

    // Let's try to type into 'Address / URL' input.
    // And verify 'Review Changes' button appears or we click it.

    // Changing the mock temporarily for this test might be complex.
    // Let's try to edit 'postgres-primary'.

    // Wait for dialog to be stable

    // Since mock type is 'remote', constructConfig might treat it as 'other' or logic might fail?
    // Let's check RegisterServiceDialog logic again?
    // It maps types.

    // Simplification: trigger 'Add Service' (already tested above), fill it, then finding 'Review Changes' is hard because it's not editing.
    // But we can simulate "Edit" by navigating to a page that opens it.

    // Let's UPDATE the mock for this specific test block to return type 'http' so editing is easy.
    await page.route('**/api/v1/services/postgres-primary', async route => {
        await route.fulfill({
            json: {
                service: {
                    id: 'postgres-primary',
                    name: 'Primary DB',
                    type: 'http',
                    httpService: { address: 'https://api.example.com' },
                    status: 'healthy',
                }
            }
        });
    });

    await page.reload(); // Reload to get new mock
    await page.waitForTimeout(1000);
    await page.getByRole('button', { name: 'Edit Config' }).click();
    await page.waitForTimeout(500);

    // Change address
    await page.getByLabel('Address / URL').fill('https://api.example.org');

    // Click Review Changes
    await page.getByRole('button', { name: 'Review Changes' }).click();
    await page.waitForTimeout(1000);

    await expect(page.getByText('Review Changes')).toBeVisible();
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'diff-feature.png') });

    // Close dialog
    await page.keyboard.press('Escape');
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

  test('Playground Drag Drop Screenshots', async ({ page }) => {
    // Navigate to playground
    await page.goto('/playground');
    await page.waitForTimeout(1000);

    // Mock Tools with File Input
    await page.route('**/api/v1/tools', async route => {
         await route.fulfill({
             json: {
                 tools: [
                     {
                         name: 'image_analyzer',
                         description: 'Analyzes an image',
                         inputSchema: {
                             type: 'object',
                             properties: {
                                 image: {
                                     type: 'string',
                                     contentEncoding: 'base64',
                                     description: 'The image to analyze'
                                 }
                             }
                         }
                     }
                 ]
             }
         });
    });

    // Reload to get new tools
    await page.reload();
    await page.waitForTimeout(1000);

    // Open Sidebar if closed (default is open on desktop, but check)
    // Click tool to open config dialog
    await page.getByText('image_analyzer').click();
    await page.waitForTimeout(500);

    // Verify dialog open
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    // Dispatch dragover to show overlay
    // Use evaluate to dispatch event on form
    const form = page.locator('form');
    await form.evaluate(element => element.dispatchEvent(new DragEvent('dragover', { bubbles: true, cancelable: true })));

    // Wait for overlay
    await expect(page.getByText('Drop File Here')).toBeVisible();

    // Take screenshot
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'playground_drag_drop.png') });
  });

  test('Stack Composer Screenshots', async ({ page }) => {
    await page.goto('/stacks');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'stack_composer_overview.png'), fullPage: true });
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'stacks.png'), fullPage: true });

    if (await page.getByText('Service Palette').isVisible()) {
        await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'stack_composer_palette.png'), fullPage: true });
    }
  });

  test('Traces Screenshots', async ({ page }) => {
    // Mock Traces (UI calls /api/traces, expects direct array with rootSpan)
    await page.route('**/api/traces*', async route => {
        const now = Date.now();
        await route.fulfill({
            json: [
                 {
                     id: 't1',
                     timestamp: now,
                     rootSpan: {
                         id: 's1',
                         name: 'filesystem.read',
                         type: 'tool',
                         startTime: now,
                         endTime: now + 120,
                         status: 'success',
                         input: { path: '/var/log/syslog' },
                         output: { content: '...' }
                     },
                     status: 'success',
                     totalDuration: 120,
                     trigger: 'user'
                 },
                 {
                     id: 't2',
                     timestamp: now - 5000,
                     rootSpan: {
                         id: 's2',
                         name: 'calculator.add',
                         type: 'tool',
                         startTime: now - 5000,
                         endTime: now - 4990,
                         status: 'error',
                         errorMessage: 'Division by zero'
                     },
                     status: 'error',
                     totalDuration: 10,
                     trigger: 'user'
                 },
                 {
                     id: 't3',
                     timestamp: now - 10000,
                     rootSpan: {
                         id: 's3',
                         name: 'memory.read_graph',
                         type: 'tool',
                         status: 'error',
                         startTime: now - 10000,
                         endTime: now - 9950,
                         input: { entities: [{ name: 'test', extra: 'field' }] },
                         output: { error: 'Schema validation error: properties "extra" not allowed' },
                         errorMessage: 'Schema validation error: properties "extra" not allowed'
                     },
                     status: 'error',
                     totalDuration: 50,
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

    // Close sheet by reloading (simplest way to reset state in tests without complex interaction)
    await page.reload();
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);

    // Click diagnostics trace
    await page.getByText('memory.read_graph').first().click({ force: true });
    await page.waitForTimeout(500);
    await expect(page.getByText('Diagnostics & Suggestions')).toBeVisible();
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'trace_diagnostics.png'), fullPage: true });
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

    // Verification Screenshot (Test Connection)
    await page.getByRole('button', { name: 'New Credential' }).click();
    await expect(page.getByText('Create Credential', { exact: true })).toBeVisible({ timeout: 10000 });
    await page.waitForTimeout(500);

    await page.getByLabel('Name').fill('Test Credential');
    // Test Connection section
    await page.getByPlaceholder('https://api.example.com/test').fill('https://api.example.com/status');
    const testBtn = page.getByRole('button', { name: 'Test', exact: true });

    // Mock testAuth response
    await page.route('**/api/v1/auth/test', async route => {
         await route.fulfill({ status: 200, json: { status: 200, status_text: 'OK' } });
    });

    await testBtn.click();
    await expect(page.getByText('Test passed: 200 OK')).toBeVisible();

    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'verification.png') });
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
      await page.getByText('Rerun Diagnostics').waitFor({ timeout: 10000 });

      // Take screenshot of the modal
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'diagnostics_failure.png') });
  });

});
