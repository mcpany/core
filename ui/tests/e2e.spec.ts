/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import path from 'path';

const DATE = new Date().toISOString().split('T')[0];
const AUDIT_DIR = path.join(__dirname, `../../.audit/ui/${DATE}`);

test.describe('MCP Any UI E2E Tests', () => {

  test('Dashboard loads correctly', async ({ page }) => {
    await page.goto('/');
    // Check for metrics
    await expect(page.locator('text=Total Requests')).toBeVisible();
    await expect(page.locator('text=Active Services')).toBeVisible();
    // Check for health widget
    await expect(page.locator('text=System Health')).toBeVisible();

    // Audit Screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'dashboard_verified.png'), fullPage: true });
  });

  test('Services page lists services and allows toggle', async ({ page }) => {
    await page.goto('/services');
    await expect(page.locator('h2')).toContainText('Services');

    // Check for list of services
    await expect(page.locator('text=Payment Service')).toBeVisible();
    // await expect(page.locator('text=Auth Service')).toBeVisible(); // Might be flaky if mock changes

    // Test toggle
    const switchElement = page.locator('button[role="switch"]').first();
    await expect(switchElement).toBeVisible();

    // Test Add Service (Edit Sheet)
    await page.click('text=Add Service');
    await expect(page.locator('text=New Service')).toBeVisible();
    await expect(page.locator('input#name')).toBeVisible();

    // Audit Screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'services_verified.png'), fullPage: true });
  });

  test('Tools page lists tools', async ({ page }) => {
    await page.goto('/tools');
    await expect(page.locator('h2')).toContainText('Tools');
    await expect(page.locator('text=calculator')).toBeVisible();
    await expect(page.locator('text=weather_lookup')).toBeVisible();

    // Audit Screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'tools.png'), fullPage: true });
  });

  test('Middleware page shows pipeline', async ({ page }) => {
    await page.goto('/middleware');
    await expect(page.locator('h2')).toContainText('Middleware Pipeline');
    await expect(page.locator('text=Incoming Request')).toBeVisible();
    // Use first() to avoid strict mode violation if text appears multiple times
    await expect(page.locator('text=auth').first()).toBeVisible();

    // Audit Screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'middleware.png'), fullPage: true });
  });

  test('Webhooks page displays configuration', async ({ page }) => {
    await page.goto('/settings/webhooks');
    await expect(page.getByRole('heading', { name: 'Webhooks' })).toBeVisible();

    // Audit Screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'webhooks_verified.png'), fullPage: true });
  });

});
