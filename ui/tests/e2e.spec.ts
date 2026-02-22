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
    await seedUser(request, "e2e-admin-core");

      // Login before each test
      await page.goto('/login');
      // Wait for page to be fully loaded as it might be transitioning
      await page.waitForLoadState('domcontentloaded');

    await page.fill('input[name="username"]', 'e2e-admin-core');
      await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]', { force: true });

      // Wait for redirect to home page and verify
    await page.waitForURL('/', { timeout: 30000 });
      await expect(page).toHaveURL('/', { timeout: 15000 });
  });
  test.afterEach(async ({ request }) => {
      await cleanupServices(request);
      await cleanupTemplates(request);
      await cleanupWebhooks(request);
    // await cleanupUser(request, "e2e-admin-core");
  });

  test('Dashboard loads correctly', async ({ page }) => {
    // Ensure System Health widget is visible
    const systemHealthCard = page.getByText('System Health').first();
    if (!(await systemHealthCard.isVisible())) {
      await page.getByTestId('add-widget-trigger').first().click();
      await page.getByText('Metrics Overview').first().click();
      await expect(systemHealthCard).toBeVisible({ timeout: 30000 });
    }

    // Check for metrics
    await expect(page.getByText('Total Requests').first()).toBeVisible({ timeout: 60000 });
    await expect(page.getByText('Active Services').first()).toBeVisible({ timeout: 60000 });
    // Check for health widget specifically
    await expect(systemHealthCard).toBeVisible({ timeout: 60000 });

    if (process.env.CAPTURE_SCREENSHOTS === 'true') {
      await page.screenshot({ path: path.join(AUDIT_DIR, 'dashboard_verified.png'), fullPage: true });
    }
  });

  test('Tools page lists tools', async ({ page }) => {
    await page.goto('/tools');
    await expect(page.getByRole('heading', { name: 'Tools' })).toBeVisible();
    // Using retry logic because tools might take a moment to sync from seeded services
    await expect(async () => {
      await page.reload();
      await expect(page.locator('text=calculator').first()).toBeVisible({ timeout: 5000 });
      await expect(page.locator('text=process_payment').first()).toBeVisible({ timeout: 5000 });
    }).toPass({ timeout: 30000, intervals: [2000, 5000] });

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
    // Check for nodes - using first() to avoid strict mode violations if multiple nodes match
    await expect(page.locator('text=Payment Gateway').first()).toBeVisible();
    await expect(page.locator('text=Math').first()).toBeVisible();

    if (process.env.CAPTURE_SCREENSHOTS === 'true') {
      await page.screenshot({ path: path.join(__dirname, 'network_topology_verified.png'), fullPage: true });
    }
  });

  test('Service Health Widget shows diagnostics', async ({ page }) => {
    await page.goto('/');

    // Ensure Service Health widget is visible
    const userService = page.locator('.group', { hasText: 'User Service' });
    if (!(await userService.isVisible())) {
      await page.getByTestId('add-widget-trigger').first().click();
      await page.getByText('Service Health').first().click();
      await expect(userService).toBeVisible({ timeout: 30000 });
    }

    await expect(userService).toBeVisible();

    // We skip checking error details as it depends on runtime health check timing
  });

});
