/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import path from 'path';
import fs from 'fs';

// Update to save directly to docs/screenshots as requested
const AUDIT_DIR = path.join(process.cwd(), 'docs/screenshots');

// Ensure audit dir exists
if (!fs.existsSync(AUDIT_DIR)){
    fs.mkdirSync(AUDIT_DIR, { recursive: true });
}

test.describe.skip('Audit Screenshots', () => {
  test.use({ colorScheme: 'dark' });

  test.beforeEach(async ({ page }) => {
    await page.addStyleTag({ content: ':root { color-scheme: dark; }' });
    await page.evaluate(() => document.documentElement.classList.add('dark'));
  });

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

  test('Capture Logs', async ({ page }) => {
    await page.goto('/logs');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(AUDIT_DIR, 'logs.png'), fullPage: true });
  });

  test('Capture Playground', async ({ page }) => {
    await page.goto('/playground');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(AUDIT_DIR, 'playground.png'), fullPage: true });
  });

  test('Capture Settings', async ({ page }) => {
    await page.goto('/settings');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(AUDIT_DIR, 'settings.png'), fullPage: true });
  });

  test('Capture Stacks', async ({ page }) => {
    await page.goto('/stacks');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(AUDIT_DIR, 'stacks.png'), fullPage: true });
  });

  test('Capture Stats', async ({ page }) => {
    await page.goto('/stats');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(AUDIT_DIR, 'stats.png'), fullPage: true });
  });

  test('Capture Network', async ({ page }) => {
    // Mock the topology endpoint to ensure the graph renders populated data
    await page.route('/api/topology', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          core: {
            id: 'core-1',
            label: 'MCP Core',
            type: 'NODE_TYPE_CORE',
            status: 'NODE_STATUS_ACTIVE',
            metrics: { qps: 12.5, latencyMs: 45, errorRate: 0.001 },
            children: [
              {
                id: 'service-weather',
                label: 'Weather Service',
                type: 'NODE_TYPE_SERVICE',
                status: 'NODE_STATUS_ACTIVE',
                metrics: { qps: 5.2, latencyMs: 120, errorRate: 0 },
                children: [
                  {
                    id: 'tool-get-weather',
                    label: 'get_weather',
                    type: 'NODE_TYPE_TOOL',
                    status: 'NODE_STATUS_ACTIVE'
                  }
                ]
              },
              {
                id: 'service-calc',
                label: 'Calculator',
                type: 'NODE_TYPE_SERVICE',
                status: 'NODE_STATUS_INACTIVE',
                children: []
              }
            ]
          },
          clients: [
            {
              id: 'client-web',
              label: 'Web Dashboard',
              type: 'NODE_TYPE_CLIENT',
              status: 'NODE_STATUS_ACTIVE',
              metrics: { qps: 2.1, latencyMs: 10 }
            },
            {
              id: 'client-cli',
              label: 'Gemini CLI',
              type: 'NODE_TYPE_CLIENT',
              status: 'NODE_STATUS_ACTIVE',
              metrics: { qps: 0.5, latencyMs: 80 }
            }
          ]
        })
      });
    });

    await page.goto('/network');
    // Wait for the graph to render (give it a bit more time for layout)
    await page.waitForTimeout(2000);
    await page.screenshot({ path: path.join(AUDIT_DIR, 'network.png'), fullPage: true });
  });
});
