/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Alerts Management', () => {
  test('should display alerts page and create a rule', async ({ page }) => {
    // 1. Navigate to Alerts Page
    await page.goto('/alerts');
    await expect(page.getByRole('heading', { name: 'Alerts' })).toBeVisible();

    // 2. Verify Stats Components are visible
    // The cards have titles: Active Critical, Active Warnings, MTTR (Today), Total Incidents
    // Use .first() or more specific locators if multiple elements match
    await expect(page.getByText('Active Critical').first()).toBeVisible();
    await expect(page.getByText('Active Warnings').first()).toBeVisible();
    await expect(page.getByText('Total Incidents').first()).toBeVisible();

    // 3. Open Create Rule Dialog
    // "Create Rule" is inside a dialog trigger in a flex container.
    // It might be getting confused or not clickable if something overlaps.
    // Try finding it by text or icon?
    // The button has "New Alert Rule" text in the component source.
    await page.getByRole('button', { name: 'New Alert Rule' }).click();
    await expect(page.getByRole('heading', { name: 'Create Alert Rule' })).toBeVisible();

    // 4. Fill Form
    await page.getByLabel('Name').fill('High Error Rate');
    await page.getByLabel('Metric').fill('error_rate');
    // Threshold is a placeholder in the input component, label mapping might be ambiguous
    await page.getByPlaceholder('Threshold').fill('5');
    await page.getByLabel('Duration').fill('10m');
    // Severity select trigger might not be associated with label correctly via aria-labelledby if not set up
    // Or it might be finding the SelectValue which is not clickable in the way we expect?
    // Use the role `combobox` which is standard for Select trigger in Radix UI
    // There are 2 comboboxes (Severity and Operator). We want Severity which defaults to "Warning".
    // Or we can find the one that contains "Warning" or "Select severity"
    await page.getByRole('combobox').filter({ hasText: 'Warning' }).click();
    // Select item from dropdown (radix-ui)
    await page.getByRole('option', { name: 'Critical' }).click();

    // 5. Submit
    // We mock the API response or rely on the backend.
    // Since we don't have a real alerts backend logic that persists rules in this mocked environment perfectly,
    // we assume the UI handles the success.
    // However, in E2E with real backend, this should work.
    // If backend doesn't support creating rules yet (mock), this might fail if we expect it to show up.
    // For now, let's verify the form submission closes the dialog.

    // Intercept the request to ensure it's sent
    const requestPromise = page.waitForRequest(request =>
        request.url().includes('/api/v1/alerts/rules') && request.method() === 'POST'
    );

    await page.getByRole('button', { name: 'Create Rule' }).click();

    const request = await requestPromise;
    expect(request.postDataJSON()).toMatchObject({
        name: 'High Error Rate',
        metric: 'error_rate',
        threshold: 5,
        duration: '10m', // The UI formats this
        severity: 'critical'
    });

    // Dialog should close
    await expect(page.getByRole('heading', { name: 'Create Alert Rule' })).not.toBeVisible();
  });
});
