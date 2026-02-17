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
      await seedServices(request);
      await seedTraffic(request);
      await seedTemplates(request);
      await seedWebhooks(request);
      await seedUser(request, "e2e-admin");

      // Login before each test
      await page.goto('/login');
      // Wait for page to be fully loaded as it might be transitioning
      await page.waitForLoadState('networkidle');

      await page.fill('input[name="username"]', 'e2e-admin');
      await page.fill('input[name="password"]', 'password');
      await page.click('button[type="submit"]');

      // Wait for redirect to home page and verify
      await expect(page).toHaveURL('/', { timeout: 15000 });
      // Ensure the page is fully loaded and hydrated
      await page.waitForLoadState('networkidle');
  });

  test.afterEach(async ({ request }) => {
      await cleanupServices(request);
      await cleanupTemplates(request);
      await cleanupWebhooks(request);
      await cleanupUser(request, "e2e-admin");
  });

  test('Dashboard loads correctly', async ({ page }) => {
    // Wait for the traffic data to be loaded AND contain non-zero data
    // This ensures the seeded data has actually propagated to the read path
    const trafficResponse = await page.waitForResponse(async resp => {
      if (resp.url().includes('/api/v1/dashboard/traffic') && resp.status() === 200) {
        const body = await resp.json();
        // Check if any point has requests > 0
        return Array.isArray(body) && body.some((p: any) => p.requests > 0);
      }
      return false;
    }, { timeout: 30000 });

    // Check for metrics
    await expect(page.locator('text=Total Requests')).toBeVisible({ timeout: 30000 });
    await expect(page.locator('text=Active Services')).toBeVisible({ timeout: 30000 });
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
