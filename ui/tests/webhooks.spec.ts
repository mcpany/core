/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import { seedUser, cleanupUser } from './e2e/test-data';

const DATE = new Date().toISOString().split('T')[0];
const AUDIT_DIR = path.join(__dirname, `../../.audit/ui/${DATE}`);

test.describe('Webhooks Page', () => {
  test.beforeEach(async ({ request, page }) => {
      await seedUser(request, "e2e-admin");

      // Login
      await page.goto('/login');
      await page.waitForLoadState('networkidle');
      await page.fill('input[name="username"]', 'e2e-admin');
      await page.fill('input[name="password"]', 'password');
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL('/', { timeout: 15000 });
  });

  test.afterEach(async ({ request }) => {
      await cleanupUser(request, "e2e-admin");
  });

  test('Lists seeded webhooks and allows creation', async ({ page }) => {
    await page.goto('/webhooks');
    await expect(page.locator('h1')).toContainText('Webhooks');

    // Verify seeded webhook
    await expect(page.locator('text=https://example.com/webhook')).toBeVisible();
    await expect(page.locator('text=Active').first()).toBeVisible();

    // Create new webhook
    await page.click('button:has-text("New Webhook")');
    await page.fill('input[id="url"]', 'https://test.local/hook');
    await page.click('button:has-text("Add Webhook")');

    // Verify new webhook
    await expect(page.locator('text=https://test.local/hook')).toBeVisible();

    if (process.env.CAPTURE_SCREENSHOTS === 'true') {
      await page.screenshot({ path: path.join(AUDIT_DIR, 'webhooks_e2e.png'), fullPage: true });
    }
  });
});
