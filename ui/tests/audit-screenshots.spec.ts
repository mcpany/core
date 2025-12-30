/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import fs from 'fs';

const DATE = new Date().toISOString().split('T')[0];
const AUDIT_DIR = path.join(process.cwd(), `.audit/ui/${DATE}`);

// Ensure audit dir exists
if (!fs.existsSync(AUDIT_DIR)){
    fs.mkdirSync(AUDIT_DIR, { recursive: true });
}

test.describe('Audit Screenshots', () => {

  test('Capture Dashboard', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000); // Wait for animations
    await page.screenshot({ path: path.join(AUDIT_DIR, 'dashboard.png'), fullPage: true });
  });

  test('Capture Services', async ({ page }) => {
    await page.goto('/services');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(AUDIT_DIR, 'services.png'), fullPage: true });
  });

  test('Capture Tools', async ({ page }) => {
    await page.goto('/tools');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(AUDIT_DIR, 'tools.png'), fullPage: true });
  });

  test('Capture Resources', async ({ page }) => {
      await page.goto('/resources');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(AUDIT_DIR, 'resources.png'), fullPage: true });
  });

  test('Capture Prompts', async ({ page }) => {
      await page.goto('/prompts');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(AUDIT_DIR, 'prompts.png'), fullPage: true });
  });

  test('Capture Profiles', async ({ page }) => {
      await page.goto('/profiles');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(AUDIT_DIR, 'profiles.png'), fullPage: true });
  });

  test('Capture Middleware', async ({ page }) => {
    await page.goto('/middleware');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(AUDIT_DIR, 'middleware.png'), fullPage: true });
  });

  test('Capture Webhooks', async ({ page }) => {
    await page.goto('/webhooks');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(AUDIT_DIR, 'webhooks.png'), fullPage: true });
  });
});
