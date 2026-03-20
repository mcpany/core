cat << 'INNER_EOF' > ui/tests/e2e/settings.spec.ts
/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Settings & Secrets', () => {
  test.beforeEach(async ({ request, page }) => {
    // Use Database Seeding instead of Mocking! (Real Data Law)
    const response = await request.post('/api/v1/debug/seed', {
      data: {
        settings: {
          mcp_listen_address: ":8080",
          log_level: 1, // INFO
          log_format: 1, // TEXT
          audit: { enabled: true },
          dlp: { enabled: false },
          gc_settings: { interval: "1h" }
        },
        secrets: [
          {
            id: "sec-1",
            name: "OpenAI Prod",
            key: "OPENAI_API_KEY",
            provider: "openai",
            created_at: new Date().toISOString(),
            value: "sk-real-data-test"
          },
          {
            id: "sec-2",
            name: "Anthropic Dev",
            key: "ANTHROPIC_API_KEY",
            provider: "anthropic",
            created_at: new Date().toISOString(),
            value: "sk-ant-test"
          }
        ]
      }
    });
    expect(response.ok()).toBeTruthy();

    await page.goto('/settings');
  });

  test('should manage global settings via backend', async ({ page }) => {
    // "General" was renamed to "Global Config"
    await page.getByRole('tab', { name: 'Global Config' }).click();

    // Switch Log Level
    const logLevelTrigger = page.getByRole('combobox').first();
    await expect(logLevelTrigger).toBeVisible();
    await logLevelTrigger.click();
    await page.getByRole('option', { name: 'DEBUG' }).click();

    // Save Settings
    await page.getByRole('button', { name: 'Save Settings' }).click();

    // Verify it saved (we can reload the page and verify state persists)
    await page.reload();
    await page.getByRole('tab', { name: 'Global Config' }).click();
    await expect(page.getByRole('combobox').first()).toHaveText('DEBUG');
  });

  test('should bulk delete secrets', async ({ page }) => {
    // Navigate to Secrets Manager tab
    await page.getByRole('tab', { name: 'Secrets & Keys' }).click();

    // Verify seeded secrets are present
    await expect(page.getByText('OpenAI Prod')).toBeVisible();
    await expect(page.getByText('Anthropic Dev')).toBeVisible();

    // Select All
    await page.getByRole('checkbox', { name: 'Select all' }).click();

    // Verify "2 selected" is shown
    await expect(page.getByText('2 selected')).toBeVisible();

    // Setup dialog handler to automatically accept confirmation
    page.on('dialog', dialog => dialog.accept());

    // Click "Delete Selected"
    await page.getByRole('button', { name: 'Delete Selected' }).click();

    // Wait for the deletion to process
    await expect(page.getByText('Secrets Deleted')).toBeVisible();

    // Verify secrets are gone
    await expect(page.getByText('No secrets found.')).toBeVisible();
  });
});
INNER_EOF
