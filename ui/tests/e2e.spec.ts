/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import path from 'path';

const DATE = new Date().toISOString().split('T')[0];
const AUDIT_DIR = path.join(__dirname, `../../.audit/ui/${DATE}`);

test.describe('MCP Any UI Overhaul', () => {
  test('Dashboard loads and displays metrics', async ({ page }) => {
    await page.goto('http://localhost:9002');
    await expect(page.getByText('Total Requests')).toBeVisible();
    await expect(page.getByText('Service Health')).toBeVisible();

    // Audit Screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'dashboard.png'), fullPage: true });
  });

  test('Services page lists services and allows toggling', async ({ page }) => {
    await page.goto('http://localhost:9002/services');
    await expect(page.getByText('Payment Gateway')).toBeVisible();

    // Check toggle
    const toggle = page.getByRole('switch').first();
    await expect(toggle).toBeVisible();

    // Open Edit Sheet
    // Use a more specific locator if the button doesn't have text
    // The button has <Settings /> icon inside
    await page.locator('button:has(.lucide-settings)').first().click();

    await expect(page.getByText('Edit Service')).toBeVisible();

    // Audit Screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'services.png'), fullPage: true });
  });

  test('Middleware page displays pipeline', async ({ page }) => {
    await page.goto('http://localhost:9002/settings/middleware');
    await expect(page.getByText('Middleware Pipeline')).toBeVisible();
    await expect(page.getByText('Global Rate Limiter')).toBeVisible();

    // Audit Screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'middleware.png'), fullPage: true });
  });

  test('Webhooks page displays configuration', async ({ page }) => {
    await page.goto('http://localhost:9002/settings/webhooks');
    // Use more specific locator to avoid ambiguity
    await expect(page.getByRole('heading', { name: 'Webhooks' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Add Webhook' })).toBeVisible();

    // Audit Screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'webhooks.png'), fullPage: true });
  });
});
