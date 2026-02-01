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
    await page.goto(`/service/${serviceName}`);

    // 2. Go to Safety Tab
    await page.getByRole('tab', { name: 'Safety' }).click();

    // 3. Verify Policy Editor is present
    await expect(page.getByText('Advanced Call Policies')).toBeVisible();
    await expect(page.getByText('No policies configured')).toBeVisible();

    // 4. Add Policy
    await page.getByRole('button', { name: 'Add Policy' }).click();

    // 5. Configure Policy (Deny by default)
    await expect(page.getByRole('dialog')).toBeVisible();

    // Select Default Action.
    // Shadcn select trigger usually has role 'combobox'.
    // We have "Default Action" label right before it.
    await page.getByRole('combobox').first().click();
    await page.getByRole('option', { name: 'Deny' }).click();

    // 6. Add Rule (Allow git)
    await page.getByRole('button', { name: 'Add Rule' }).click();

    // By default newly added rule is Allow.
    // Fill Name Regex
    await page.getByPlaceholder('e.g. ^git.*').fill('^git.*');

    // 7. Save
    await page.getByRole('button', { name: 'Save Policy' }).click();
    await expect(page.getByRole('dialog')).not.toBeVisible();

    // 8. Verify UI update
    // The DashboardGrid uses "Deny" label logic from our component
    // ACTION_LABELS[1] is "Deny".
    // However, our component renders badges or text.
    // Looking at code: <span className="text-xs font-bold mt-1">{ACTION_LABELS[policy.defaultAction]}</span>
    await expect(page.getByText('Deny').first()).toBeVisible();
    await expect(page.getByText('Name: /^git.*/')).toBeVisible();

    // 9. Verify Toast (Policies Updated)
    await expect(page.getByText('Policies Updated').first()).toBeVisible();

    // 10. Verify Backend Persistence
    // Reload page to ensure it's not just local state
    await page.reload();
    await page.getByRole('tab', { name: 'Safety' }).click();
    await expect(page.getByText('Deny').first()).toBeVisible();
    await expect(page.getByText('Name: /^git.*/')).toBeVisible();

    // Check API directly
    const response = await request.get(`/api/v1/services/${serviceName}`);
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    // API returns the object directly, or wrapped depending on implementation.
    // Based on debug logs, it returns the object directly.
    const policies = data.call_policies || data.service?.call_policies;
    expect(policies).toBeDefined();
    expect(policies).toHaveLength(1);
    // Enums are returned as strings in JSON
    expect(policies[0].default_action).toBe("DENY");
    expect(policies[0].rules).toHaveLength(1);
    expect(policies[0].rules[0].name_regex).toBe('^git.*');
  });
});
