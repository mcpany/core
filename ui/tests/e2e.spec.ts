/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import { seedServices, seedTraffic, cleanupServices, seedUser, cleanupUser } from './e2e/test-data';

const DATE = new Date().toISOString().split('T')[0];
const AUDIT_DIR = path.join(__dirname, `../../.audit/ui/${DATE}`);

test.describe('MCP Any UI E2E Tests', () => {
  test.describe.configure({ mode: 'serial' });

  test.beforeEach(async ({ request, page }) => {
      await seedServices(request);
      await seedTraffic(request);
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
  });

  test.afterEach(async ({ request }) => {
      await cleanupServices(request);
      await cleanupUser(request, "e2e-admin");
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
    // Check if the service name is visible in the Service Health widget
    await expect(page.locator('text=User Service').first()).toBeVisible();
  });

});
