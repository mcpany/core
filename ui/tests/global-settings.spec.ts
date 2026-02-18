/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Global Security Settings', () => {
  test.beforeEach(async ({ request }) => {
    // Reset global settings to a known state
    // We assume default state is clean enough, but explicit set is better
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
    // 1. Navigate to Settings
    await page.goto('/settings');

    // 2. Go to Security Tab
    await page.getByRole('tab', { name: 'Security & Governance' }).click();

    // 3. Enable DLP
    const dlpSwitch = page.getByLabel('Enable DLP');
    await expect(dlpSwitch).toBeVisible();
    // If already enabled (shouldn't be due to beforeEach), don't click?
    // beforeEach sets enabled: false.
    await dlpSwitch.click();

    // 4. Add Custom Pattern
    await page.getByRole('button', { name: 'Add Pattern' }).click();
    await page.getByPlaceholder('e.g. sk-[a-zA-Z0-9]+').fill('sk-test-[0-9]+');

    // 5. Configure Rate Limit
    const rateLimitSwitch = page.getByLabel('Enable Rate Limiting');
    await rateLimitSwitch.click();

    const rpsInput = page.getByLabel('Requests Per Second (RPS)');
    await rpsInput.fill('100');

    // 6. Save
    await page.getByRole('button', { name: 'Save Security Settings' }).click();

    // 7. Verify Toast
    await expect(page.getByText('Security settings saved')).toBeVisible();

    // 8. Verify Backend Persistence
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

    // Verify Audit is enabled (from beforeEach)
    const auditSwitch = page.getByLabel('Enable Audit Logging');
    await expect(auditSwitch).toBeChecked();

    // Toggle Log Arguments
    const logArgs = page.getByLabel('Log Arguments');
    await logArgs.click(); // Enable it

    // Change Storage to SQLite (Select)
    // shadcn select trigger
    await page.getByRole('combobox').filter({ hasText: 'Storage Backend' }).click();
    // Or try finding by label mapping if possible, but filter is robust.
    // Alternatively: page.locator('button[role="combobox"]').nth(0) if order is known?
    // Let's use the text inside the trigger if visible, usually "Local File".
    // Or better:
    // The label "Storage Backend" points to the select trigger?
    // Shadcn Select trigger is a button.
    // await page.getByLabel('Storage Backend').click(); // Might not work if label points to hidden select

    // Fallback: Click the button that contains "Local File" (default)
    // await page.getByRole('combobox').click(); // might be ambiguous if multiple selects

    // Look at component:
    // <FormField name="audit_storage" ... <FormLabel>Storage Backend</FormLabel> ... <Select ...>
    // There is only one Select in this card visible initially? No, duplicates hidden?
    // There is one select for Storage Backend.
    // Let's try:
    await page.getByText('Local File').click(); // Click the Trigger value
    await page.getByRole('option', { name: 'SQLite' }).click();

    // Save
    await page.getByRole('button', { name: 'Save Security Settings' }).click();
    await expect(page.getByText('Security settings saved')).toBeVisible();

    // Verify
    const response = await request.get('/api/v1/settings');
    const settings = await response.json();

    expect(settings.audit.log_arguments).toBe(true);
    expect(settings.audit.storage_type).toBe("STORAGE_TYPE_SQLITE");
  });
});
