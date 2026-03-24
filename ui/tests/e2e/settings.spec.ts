/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Settings & Secrets', () => {
  test.beforeEach(async ({ page }) => {
    // Mock Global Settings API
    await page.route(/\/api\/v1\/settings/, async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          json: {
            mcp_listen_address: ":8080",
            log_level: 1, // INFO
            log_format: 1, // TEXT
            audit: { enabled: true },
            dlp: { enabled: false },
            gc_settings: { interval: "1h" },
            read_only: false
          }
        });
      } else if (route.request().method() === 'POST') {
        await route.fulfill({ status: 200, json: {} });
      } else {
        await route.continue();
      }
    });

    // Mock other noisy endpoints
    await page.route(/\/api\/v1\/doctor/, async route => {
        await route.fulfill({ status: 200, json: { status: "ok", checks: {} } });
    });
    await page.route(/\/api\/v1\/topology/, async route => {
        await route.fulfill({ status: 200, json: { nodes: [], edges: [] } });
    });

    // Mock Secrets API with state
    const secretsStore: any[] = [];

    await page.route(/\/api\/v1\/secrets/, async route => {
      const method = route.request().method();

      if (method === 'GET') {
        await route.fulfill({
          json: { secrets: [...secretsStore] }
        });
      } else if (method === 'POST') {
        const data = route.request().postDataJSON();
        const newSecret = {
            id: `sec-${Date.now()}`,
            name: data.name,
            key: data.key,
            provider: data.provider,
            createdAt: new Date().toISOString(),
            lastUsed: null,
            value: '********' // masked
        };
        secretsStore.push(newSecret);

        await route.fulfill({
          status: 201,
          json: newSecret
        });
      } else if (method === 'DELETE') {
        await route.fulfill({ status: 200, json: {} });
      } else {
        await route.continue();
      }
    });

    await page.route(/\/api\/v1\/secrets\/.+/, async route => {
        if (route.request().method() === 'DELETE') {
            const url = route.request().url();
            const id = url.split('/').pop();
            const index = secretsStore.findIndex(s => s.id === id);
            if (index !== -1) {
                secretsStore.splice(index, 1);
            }
            await route.fulfill({ status: 200, json: {} });
        } else {
            await route.continue();
        }
    });

    await page.goto('/settings');
  });

  test('should manage global settings', async ({ page }) => {
    // Global Settings (Log Level)
    // "General" was renamed to "Global Config"
    await page.getByRole('tab', { name: 'Global Config' }).click();
    const logLevelTrigger = page.getByRole('combobox').first();
    await expect(logLevelTrigger).toBeVisible();
    await logLevelTrigger.click();
    await page.getByRole('option', { name: 'DEBUG' }).click();
    await page.getByRole('button', { name: 'Save Settings' }).click();
  });
});
