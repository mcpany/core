/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import path from 'path';

// Use current date for audit directory
const date = new Date().toISOString().split('T')[0];
const AUDIT_DIR = path.join(__dirname, `../.audit/ui/${date}`);

test.describe('Audit Screenshots', () => {

  test('Capture Dashboard', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(AUDIT_DIR, 'dashboard.png'), fullPage: true });
  });

  test('Capture Services', async ({ page }) => {
    await page.goto('/services');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(AUDIT_DIR, 'services.png'), fullPage: true });
  });

  test('Capture Tools', async ({ page }) => {
    await page.goto('/tools');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(AUDIT_DIR, 'tools.png'), fullPage: true });
  });

  test('Capture Resources', async ({ page }) => {
    await page.goto('/resources');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(AUDIT_DIR, 'resources.png'), fullPage: true });
  });

  test('Capture Prompts', async ({ page }) => {
    await page.goto('/prompts');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(AUDIT_DIR, 'prompts.png'), fullPage: true });
  });

  test('Capture Profiles', async ({ page }) => {
    await page.goto('/dashboard/profiles');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(AUDIT_DIR, 'profiles.png'), fullPage: true });
  });

  test('Capture Middleware', async ({ page }) => {
    await page.goto('/dashboard/middleware');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(AUDIT_DIR, 'middleware.png'), fullPage: true });
  });

    test('Capture Webhooks', async ({ page }) => {
    await page.goto('/dashboard/webhooks');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(AUDIT_DIR, 'webhooks.png'), fullPage: true });
  });

    test('Capture Settings', async ({ page }) => {
    await page.goto('/settings');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(AUDIT_DIR, 'settings.png'), fullPage: true });
  });
});
