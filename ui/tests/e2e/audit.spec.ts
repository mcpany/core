/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';

test.describe('Audit Smoke Tests', () => {
    // Instead of screenshots, we ensure critical elements are present on all key pages.
  test('should verify all main pages load correctly', async ({ page }) => {
    const pages = [
      { url: '/', check: 'Dashboard' },
      { url: '/services', check: 'Services' },
      { url: '/tools', check: 'Tools' },
      { url: '/resources', check: 'Resources' },
      { url: '/prompts', check: 'Prompts' },
      { url: '/profiles', check: 'Profiles' },
      { url: '/middleware', check: 'Middleware Pipeline' },
      { url: '/webhooks', check: 'Webhooks' },
    ];

    for (const p of pages) {
      await page.goto(p.url);
      await expect(page.locator('body')).toContainText(p.check);
    }

    // Corner case: 404
    await page.goto('/non-existent-page');
    await expect(page.locator('body')).toContainText('404');
  });
});

test.describe('MCP Any Audit Screenshots', () => {
  // Skip these tests unless explicitly requested to avoid polluting git history
  test.skip(process.env.CAPTURE_SCREENSHOTS !== 'true', 'Skipping audit screenshots');

  const date = new Date().toISOString().split('T')[0];
  const auditDir = path.join(__dirname, '../../.audit/ui', date);

  test.beforeAll(async () => {
    if (!fs.existsSync(auditDir)) {
      fs.mkdirSync(auditDir, { recursive: true });
    }
  });

  test('Capture Services', async ({ page }) => {
    await page.goto('/services');
    await page.waitForSelector('text=Upstream Services');
    // Wait for either list headers (if services exist) or empty state message
    await page.waitForTimeout(1000); 
    await page.screenshot({ path: path.join(auditDir, 'services.png'), fullPage: true });
  });

  test('Capture Tools', async ({ page }) => {
     await page.goto('/tools');
     await page.waitForSelector('h1:has-text("Tools")');
     await page.waitForTimeout(1000);
     await page.screenshot({ path: path.join(auditDir, 'tools.png'), fullPage: true });
  });

  test('Capture Resources', async ({ page }) => {
     await page.goto('/resources');
     await page.waitForSelector('h1:has-text("Resources")');
     await page.waitForTimeout(1000);
     await page.screenshot({ path: path.join(auditDir, 'resources.png'), fullPage: true });
  });

  test('Capture Prompts', async ({ page }) => {
     await page.goto('/prompts');
     await page.waitForSelector('h3:has-text("Prompt Library")');
     await page.waitForTimeout(1000);
     await page.screenshot({ path: path.join(auditDir, 'prompts.png'), fullPage: true });
  });
});
