/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

test.describe('MCP Any Audit Screenshots', () => {

  const date = new Date().toISOString().split('T')[0];
  const auditDir = path.join(process.cwd(), '..', '.audit/ui', date);

  test.beforeAll(async () => {
    if (!fs.existsSync(auditDir)) {
      fs.mkdirSync(auditDir, { recursive: true });
    }
  });

  test('Capture Dashboard', async ({ page }) => {
    await page.goto('/');
    // Wait for metrics to load
    await page.waitForSelector('text=Total Requests', { timeout: 10000 });
    // Wait for health widget to load
    await page.waitForSelector('text=Core API Gateway', { timeout: 10000 });
    await page.waitForTimeout(500); // Slight buffer for animations
    await page.screenshot({ path: path.join(auditDir, 'dashboard.png'), fullPage: true });
  });

  test('Capture Services', async ({ page }) => {
    await page.goto('/services');
    await page.waitForSelector('text=Upstream Services');
    await page.waitForSelector('text=weather-service'); // Wait for list
    await page.screenshot({ path: path.join(auditDir, 'services.png'), fullPage: true });
  });

  test('Capture Tools', async ({ page }) => {
    await page.goto('/tools');
    await page.waitForSelector('text=Available Tools');
    await page.waitForSelector('text=get_weather'); // Wait for list
    await page.screenshot({ path: path.join(auditDir, 'tools.png'), fullPage: true });
  });

  test('Capture Resources', async ({ page }) => {
    await page.goto('/resources');
    await page.waitForSelector('text=Managed Resources');
    await page.waitForSelector('text=System Logs'); // Wait for list
    await page.screenshot({ path: path.join(auditDir, 'resources.png'), fullPage: true });
  });

  test('Capture Prompts', async ({ page }) => {
    await page.goto('/prompts');
    await page.waitForSelector('text=Prompt Templates');
    await page.waitForSelector('text=summarize_text'); // Wait for list
    await page.screenshot({ path: path.join(auditDir, 'prompts.png'), fullPage: true });
  });

  test('Capture Profiles', async ({ page }) => {
    await page.goto('/profiles');
    await page.waitForSelector('text=Profiles');
    await page.waitForSelector('text=Default Dev');
    await page.screenshot({ path: path.join(auditDir, 'profiles.png'), fullPage: true });
  });

  test('Capture Middleware', async ({ page }) => {
    await page.goto('/middleware');
    await page.waitForSelector('text=Middleware Pipeline');
    await page.waitForSelector('text=Authentication');
    await page.screenshot({ path: path.join(auditDir, 'middleware.png'), fullPage: true });
  });

  test('Capture Webhooks', async ({ page }) => {
    await page.goto('/webhooks');
    await page.waitForSelector('text=Webhooks');
    await page.waitForSelector('text=Configured Webhooks');
    await page.screenshot({ path: path.join(auditDir, 'webhooks.png'), fullPage: true });
  });

});
