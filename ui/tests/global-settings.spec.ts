/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Global Security Settings', () => {
  test.beforeEach(async ({ request }) => {
    // Reset global settings to a known state
    await request.post('/api/v1/settings', {
      data: {
        dlp: {
            enabled: false,
            custom_patterns: []
        },
        rate_limit: {
            is_enabled: false,
            requests_per_second: 20,
            burst: 50
        },
        audit: {
            enabled: true,
            storage_type: "STORAGE_TYPE_FILE",
            output_path: "audit.log",
            log_arguments: false,
            log_results: false
        }
      }
    });
  });

  test('should configure DLP patterns and rate limits', async ({ page, request }) => {
    await page.goto('/settings');
    await page.getByRole('tab', { name: 'Security & Governance' }).click();

    // Enable DLP
    const dlpSwitch = page.getByLabel('Enable DLP');
    await expect(dlpSwitch).toBeVisible();
    await dlpSwitch.click();

    // Add Custom Pattern
    await page.getByRole('button', { name: 'Add Pattern' }).click();
    await page.getByPlaceholder('e.g. sk-[a-zA-Z0-9]+').fill('sk-test-[0-9]+');

    // Configure Rate Limit
    const rateLimitSwitch = page.getByLabel('Enable Rate Limiting');
    await rateLimitSwitch.click();

    const rpsInput = page.getByLabel('Requests Per Second (RPS)');
    await rpsInput.fill('100');

    // Save and Wait for Response
    const savePromise = page.waitForResponse(resp => resp.url().includes('/api/v1/settings') && resp.status() === 200);
    await page.getByRole('button', { name: 'Save Security Settings' }).click();
    await savePromise;

    // Verify Backend Persistence
    const response = await request.get('/api/v1/settings');
    expect(response.ok()).toBeTruthy();
    const settings = await response.json();

    expect(settings.dlp.enabled).toBe(true);
    expect(settings.dlp.custom_patterns).toContain('sk-test-[0-9]+');
    expect(settings.rate_limit.is_enabled).toBe(true);
    expect(settings.rate_limit.requests_per_second).toBe(100);
  });

  test('should configure audit logging', async ({ page, request }) => {
    await page.goto('/settings');
    await page.getByRole('tab', { name: 'Security & Governance' }).click();

    const auditSwitch = page.getByLabel('Enable Audit Logging');
    await expect(auditSwitch).toBeChecked();

    // Toggle Log Arguments
    const logArgs = page.getByLabel('Log Arguments');
    await expect(logArgs).not.toBeChecked(); // Ensure initially unchecked
    await logArgs.click();
    await expect(logArgs).toBeChecked(); // Ensure checked after click

    // Change Storage to SQLite
    // Use the label locator
    const storageSelect = page.getByLabel('Storage Backend');
    await storageSelect.click();
    await page.getByRole('option', { name: 'SQLite' }).click();

    // Verify the select value changed (the text in the button usually updates)
    await expect(storageSelect).toHaveText('SQLite');

    // Save
    const savePromise = page.waitForResponse(resp => resp.url().includes('/api/v1/settings') && resp.status() === 200);
    await page.getByRole('button', { name: 'Save Security Settings' }).click();
    await savePromise;

    // Verify
    const response = await request.get('/api/v1/settings');
    const settings = await response.json();

    expect(settings.audit.log_arguments).toBe(true);
    // Backend enum might return string or number depending on serializer
    // protojson default is string for enums
    expect(settings.audit.storage_type).toBe("STORAGE_TYPE_SQLITE");
  });
});
