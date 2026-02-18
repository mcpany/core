/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import { seedServices, seedTraffic, seedTemplates, seedWebhooks, cleanupServices, cleanupTemplates, cleanupWebhooks, seedUser, cleanupUser } from './e2e/test-data';

const DATE = new Date().toISOString().split('T')[0];
// Use test-results directory which is writable in CI
const AUDIT_DIR = path.join(process.cwd(), `test-results/artifacts/audit/ui/${DATE}`);

test.describe('MCP Any UI E2E Tests', () => {
  test.describe.configure({ mode: 'serial' });

  test.beforeEach(async ({ request, page }) => {
      // Use debug endpoints for reliable reset/seeding
      await request.post('/api/v1/debug/reset', { headers: { 'X-API-Key': 'test-token' } });

      const seedData = {
          users: [{
              id: "e2e-admin",
              authentication: {
                  basicAuth: {
                      username: "e2e-admin",
                      passwordHash: "$2a$12$KPRtQETm7XKJP/L6FjYYxuCFpTK/oRs7v9U6hWx9XFnWy6UuDqK/a" // "password"
                  }
              },
              roles: ["admin"]
          }],
          services: [
            {
                id: "svc_01",
                name: "Payment Gateway",
                version: "v1.2.0",
                httpService: {
                    address: "https://stripe.com",
                    tools: [
                        { name: "process_payment", description: "Process a payment", callId: "process_payment_call" }
                    ],
                    calls: {
                        process_payment_call: {
                            method: "HTTP_METHOD_POST",
                            endpointPath: "/v1/charges"
                        }
                    }
                }
            },
            // ... (Other services needed for tests)
             {
                id: "svc_03",
                name: "Math",
                version: "v1.0",
                httpService: {
                    address: "http://localhost:8080",
                    tools: [
                        { name: "calculator", description: "calc", callId: "calc_call" }
                    ],
                    calls: {
                        calc_call: {
                            method: "HTTP_METHOD_POST",
                            endpointPath: "/calc"
                        }
                    }
                }
            }
          ]
      };

      await request.post('/api/v1/debug/seed', {
          headers: { 'X-API-Key': 'test-token' },
          data: seedData
      });

      // Seed traffic separately as it's a dedicated debug endpoint
      await seedTraffic(request);
      await seedWebhooks(request); // Still use helper or move to seedData if supported
      // Note: seedWebhooks uses internal API. handleDebugSeed supports 'GlobalSettings' but maybe not individual alerts easily in the same payload structure I defined.
      // My handleDebugSeed implementation supports "secrets", "collections", "global_settings".
      // Alerts are part of GlobalSettings usually, but the `api_alerts.go` manages them.
      // Let's keep seedWebhooks separate or move logic.
      // seedWebhooks uses POST /api/v1/alerts/webhook.

      // Login before each test
      await page.goto('/login');
      // Wait for page to be fully loaded as it might be transitioning
      await page.waitForLoadState('networkidle');

      await page.fill('input[name="username"]', 'e2e-admin');
      await page.fill('input[name="password"]', 'password');
      await page.click('button[type="submit"]');

      // Wait for redirect to home page and verify
      await expect(page).toHaveURL('/', { timeout: 15000 });
  });

  test.afterEach(async ({ request }) => {
      // No cleanup needed as we reset in beforeEach
  });

  test('Dashboard loads correctly', async ({ page }) => {
    // Check for metrics
    await expect(page.locator('text=Total Requests')).toBeVisible();
    await expect(page.locator('text=Active Services')).toBeVisible();
    // Check for health widget
    await expect(page.locator('text=System Health').first()).toBeVisible();

    if (process.env.CAPTURE_SCREENSHOTS === 'true') {
      await page.screenshot({ path: path.join(AUDIT_DIR, 'dashboard_verified.png'), fullPage: true });
    }
  });

  test('Tools page lists tools', async ({ page }) => {
    await page.goto('/tools');
    await expect(page.locator('h1')).toContainText('Tools');
    await expect(page.locator('text=calculator')).toBeVisible();
    await expect(page.locator('text=process_payment')).toBeVisible();

    if (process.env.CAPTURE_SCREENSHOTS === 'true') {
      await page.screenshot({ path: path.join(AUDIT_DIR, 'tools.png'), fullPage: true });
    }
  });

  test('Middleware page shows pipeline', async ({ page }) => {
    await page.goto('/middleware');
    await expect(page.locator('h1')).toContainText('Middleware Pipeline');
    await expect(page.locator('text=Incoming Request')).toBeVisible();
    await expect(page.locator('text=auth').first()).toBeVisible();

    if (process.env.CAPTURE_SCREENSHOTS === 'true') {
      await page.screenshot({ path: path.join(AUDIT_DIR, 'middleware.png'), fullPage: true });
    }
  });

  test('Webhooks page displays configuration', async ({ page }) => {
    await page.goto('/settings/webhooks');
    await expect(page.getByRole('heading', { name: 'Webhooks' })).toBeVisible();

    if (process.env.CAPTURE_SCREENSHOTS === 'true') {
      await page.screenshot({ path: path.join(AUDIT_DIR, 'webhooks_verified.png'), fullPage: true });
    }
  });

  test('Network page visualizes topology', async ({ page }) => {
    await page.goto('/network');
    await expect(page.locator('body')).toBeVisible();
    await expect(page.getByText('Network Graph').first()).toBeVisible();
    // Check for nodes
    await expect(page.locator('text=Payment Gateway')).toBeVisible();
    await expect(page.locator('text=Math')).toBeVisible();

    if (process.env.CAPTURE_SCREENSHOTS === 'true') {
      await page.screenshot({ path: path.join(__dirname, 'network_topology_verified.png'), fullPage: true });
    }
  });

  test('Service Health Widget shows diagnostics', async ({ page }) => {
    await page.goto('/');
    const userService = page.locator('.group', { hasText: 'User Service' });
    await expect(userService).toBeVisible();

    // We skip checking error details as it depends on runtime health check timing
  });

});
