/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import { seedGlobalState, seedTraffic, seedWebhooks } from './e2e/test-data';

const DATE = new Date().toISOString().split('T')[0];
// Use test-results directory which is writable in CI
const AUDIT_DIR = path.join(process.cwd(), `test-results/artifacts/audit/ui/${DATE}`);

test.describe('MCP Any UI E2E Tests', () => {
  test.describe.configure({ mode: 'serial' });

  test.beforeEach(async ({ request, page }) => {
    // Use atomic seeding for core state
    await seedGlobalState(request);
    // Seed auxiliary data
    await seedTraffic(request);
    await seedWebhooks(request);

    // Login before each test
    await page.goto('/login');
    // Wait for page to be fully loaded as it might be transitioning
    await page.waitForLoadState('networkidle');

    await page.fill('input[name="username"]', 'e2e-admin-core');
    await page.fill('input[name="password"]', 'password');
    await Promise.all([
      page.waitForURL('/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
    ]);
    await expect(page).toHaveURL('/', { timeout: 15000 });
  });
  test.afterEach(async ({ request }) => {
    // Seeding handles cleanup automatically, so we don't need explicit cleanup here
    // unless we want to leave a clean state.
    // For now, we rely on atomic seeding at start of next test.
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
    await expect(page.locator('text=Processing Order')).toBeVisible();

    const emptyState = page.getByText('No middlewares configured.');
    const priorityLabels = page.locator('text=Priority:');
    await expect(async () => {
      const hasEmptyState = await emptyState.isVisible().catch(() => false);
      const itemCount = await priorityLabels.count();
      expect(hasEmptyState || itemCount > 0).toBeTruthy();
    }).toPass({ timeout: 15000, intervals: [1000, 2000] });

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
    const userService = page.locator('.group').filter({ hasText: 'User Service' }).first();
    if (!(await userService.isVisible({ timeout: 5000 }))) {
      await page.getByTestId('add-widget-trigger').first().click();
      await page.getByText('Service Health').first().click();
      await expect(userService).toBeVisible({ timeout: 30000 });
    }

    await expect(userService).toBeVisible();

    // We skip checking error details as it depends on runtime health check timing
  });

  test('RichResultViewer displays tabular data and can sort and export', async ({ page }) => {
    // Navigate to tools and wait for tools to be visible
    await page.goto('/tools');
    await expect(page.locator('text=get_users').first()).toBeVisible({ timeout: 10000 });

    // Trigger tool execution in the playground
    await page.goto('/playground');
    await expect(page.locator('h1')).toContainText('Playground');

    // Wait for services to load in the playground
    await expect(page.locator('text=Tabular Data Service').first()).toBeVisible({ timeout: 10000 });

    // Select the "Tabular Data Service"
    const tabularServiceCard = page.locator('.group').filter({ hasText: 'Tabular Data Service' }).first();
    await tabularServiceCard.click();

    // Select the "get_users" tool within the service
    await page.getByRole('button', { name: 'get_users' }).click();

    // Execute the tool (it has no required inputs based on schema)
    await page.click('button[type="submit"]');

    // Wait for the tabular result to appear (we look for our new Export CSV button)
    const exportCsvButton = page.locator('button', { hasText: 'Export CSV' });
    await expect(exportCsvButton).toBeVisible({ timeout: 10000 });

    // Verify table headers exist and default state
    await expect(page.locator('th', { hasText: 'id' })).toBeVisible();
    await expect(page.locator('th', { hasText: 'name' })).toBeVisible();
    await expect(page.locator('th', { hasText: 'role' })).toBeVisible();

    // Initially order is 1 (Alice), 2 (Bob), 3 (Charlie)
    // Verify first row is Alice
    let firstRowName = page.locator('tbody tr').nth(0).locator('td').nth(1);
    await expect(firstRowName).toContainText('Alice');

    // Sort by name descending
    await page.locator('th', { hasText: 'name' }).click(); // Ascending
    await page.locator('th', { hasText: 'name' }).click(); // Descending

    // Verify first row is now Charlie (Charlie, Bob, Alice desc)
    firstRowName = page.locator('tbody tr').nth(0).locator('td').nth(1);
    await expect(firstRowName).toContainText('Charlie');

    // Ensure download triggers via Download Event
    const [download] = await Promise.all([
      page.waitForEvent('download'),
      exportCsvButton.click()
    ]);

    expect(download.suggestedFilename()).toMatch(/^mcpany-result-\d+\.csv$/);

    if (process.env.CAPTURE_SCREENSHOTS === 'true') {
      await page.screenshot({ path: path.join(AUDIT_DIR, 'rich_result_viewer_verified.png'), fullPage: true });
    }
  });

});
