/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Policy Editor', () => {
  const serviceName = 'policy-test-service';

  test.beforeEach(async ({ request }) => {
    // Clean up first just in case
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});

    // Seed the database
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        http_service: {
            address: "http://example.com"
        }
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterEach(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should create, update, and persist call policies', async ({ page, request }) => {
    // 1. Navigate to Service Detail
    // Note: The URL might be /upstream-services?service=... or /service/... depending on router.
    // Based on directory structure, /service/[id] exists.
    // But checking upstream-services page, it opens a sheet.
    // Let's assume /upstream-services is the entry point and we can click edit.
    // OR we can go to /upstream-services and click the row.
    // Let's try navigating to the list and clicking edit to be safe, or just use the sheet if we can trigger it.
    // Actually, `ui/src/components/services/editor/service-editor.tsx` is used inside a Sheet in `upstream-services/page.tsx`.
    // So we should go to /upstream-services, find the service row, and click Edit.

    await page.goto('/upstream-services');

    // Wait for services to load
    await expect(page.getByText(serviceName)).toBeVisible();

    // Click Edit button for the service
    // Assuming there is an edit button or clicking the row works.
    // In `service-list.tsx`, usually there is an action menu or row click.
    // Let's try finding a button with aria-label or text.
    // Usually "Edit" or an icon.
    // Let's try to click the row or "Edit" text.
    await page.getByText(serviceName).click();

    // 2. Go to Policies Tab
    // First click Settings tab (Top level)
    await page.getByRole('tab', { name: 'Settings' }).click();
    // Then click Policies tab (Nested in Editor)
    await page.getByRole('tab', { name: 'Policies' }).click();

    // 3. Verify Policy Editor is present
    await expect(page.getByText('Security & Access Control')).toBeVisible();
    await expect(page.getByText('No policies configured')).toBeVisible();

    // 4. Add Policy
    await page.getByRole('button', { name: 'Add Policy' }).click();

    // 5. Verify Policy #1 appears
    await expect(page.getByText('Policy #1')).toBeVisible();

    // 6. Expand Policy #1 (It might be expanded by default or we click the trigger)
    // The trigger contains the text "Policy #1".
    const trigger = page.getByText('Policy #1');
    await trigger.click(); // Ensure expanded

    // 7. Change Default Action to Deny All
    // By default it is Deny All. Let's toggle to Allow All then back to Deny All to verify interaction.
    // Scope to the Policy #1 container
    const policyContainer = page.locator('.border.rounded-lg').filter({ hasText: 'Policy #1' });

    // Find the combobox. Note: Shadcn Select trigger has role 'combobox'.
    const defaultActionSelect = policyContainer.getByRole('combobox').first();

    // Verify default is Deny All
    await expect(defaultActionSelect).toContainText('Deny All');

    // Change to Allow All
    await defaultActionSelect.click();
    await page.getByRole('option', { name: 'Allow All' }).click();
    await expect(defaultActionSelect).toContainText('Allow All');

    // Change back to Deny All
    await defaultActionSelect.click();
    await page.getByRole('option', { name: 'Deny All' }).click();
    await expect(defaultActionSelect).toContainText('Deny All');

    // 8. Add Rule
    await policyContainer.getByRole('button', { name: 'Add Rule' }).click();

    // 9. Configure Rule
    // Name Regex input.
    // Scope to policyContainer to avoid ambiguity if other editors have similar inputs
    await policyContainer.getByPlaceholder('e.g. ^git.*').fill('^git.*');

    // 10. Save Changes
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // 11. Verify Toast
    await expect(page.getByText('Service Updated')).toBeVisible();

    // 12. Verify Backend Persistence
    // Reload page
    await page.reload();
    // We are already on the detail page after reload usually, but let's be safe.
    // If reload keeps URL, we are at /upstream-services/policy-test-service.
    // Need to click Settings again as default is Overview.
    await page.getByRole('tab', { name: 'Settings' }).click();
    await page.getByRole('tab', { name: 'Policies' }).click();

    // Verify UI state
    await expect(page.getByText('Policy #1')).toBeVisible();
    await trigger.click();
    await expect(page.getByText('Deny All')).toBeVisible();
    // Rule input should have value
    await expect(page.getByPlaceholder('e.g. ^git.*')).toHaveValue('^git.*');

    // Check API
    const apiRes = await request.get(`/api/v1/services/${serviceName}`);
    expect(apiRes.ok()).toBeTruthy();
    const data = await apiRes.json();
    // Depending on API structure (wrapper or direct)
    const serviceData = data.service || data;
    const policies = serviceData.call_policies;

    expect(policies).toBeDefined();
    expect(policies).toHaveLength(1);
    // Enum 1 is DENY
    // Standard protojson emits strings by default.
    expect(['DENY', 1]).toContain(policies[0].default_action);
    expect(policies[0].rules).toHaveLength(1);
    expect(policies[0].rules[0].name_regex).toBe('^git.*');
  });
});
