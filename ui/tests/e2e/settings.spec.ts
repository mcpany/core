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

  test('should manage global settings and secrets', async ({ page }) => {
    // Global Settings (Log Level)
    await page.getByRole('tab', { name: 'General' }).click();
    const logLevelTrigger = page.getByRole('combobox').first();
    await expect(logLevelTrigger).toBeVisible();
    await logLevelTrigger.click();
    await page.getByRole('option', { name: 'DEBUG' }).click();
    await page.getByRole('button', { name: 'Save Settings' }).click();

    // Secrets Management
    await page.getByRole('tab', { name: 'Secrets & Keys' }).click();

    // Wait for empty state
    await expect(page.getByText('No secrets found.')).toBeVisible({ timeout: 10000 });

    await page.getByRole('button', { name: 'Add Secret' }).click();

    const secretName = `test-secret-${Date.now()}`;
    await page.fill('input[id="name"]', secretName);
    await page.fill('input[id="key"]', 'TEST_KEY');
    await page.fill('input[id="value"]', 'TEST_VAL');

    // Handle potential dialog animation
    await page.waitForTimeout(500);

    await page.getByRole('button', { name: 'Save Secret' }).click();
    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10000 });

    // Verify visibility
    // Just wait for it to appear
    await expect(page.getByText(secretName)).toBeVisible({ timeout: 10000 });

    // Verify deletion
    const secretRow = page.locator('.group').filter({ hasText: secretName });

    page.once('dialog', dialog => dialog.accept());

    await secretRow.getByLabel('Delete secret').click();

    await expect(page.getByText(secretName)).not.toBeVisible({ timeout: 10000 });
  });
});
