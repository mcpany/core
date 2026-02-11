/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Tool Visibility Policy', () => {
  const serviceName = 'visibility-test-service';

  test.beforeEach(async ({ request }) => {
    // Clean up first just in case
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});

    // Seed the database with a service that has MANUAL tools defined.
    // This allows us to test the "Quick Visibility" UI without needing a real running MCP server.
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        http_service: {
            address: "http://example.com",
            tools: [
                { name: "tool_a", description: "Tool A" },
                { name: "tool_b", description: "Tool B" },
                { name: "tool_c", description: "Tool C" }
            ]
        }
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterEach(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should toggle tool visibility using Quick Control', async ({ page, request }) => {
    // 1. Navigate to Service Detail
    await page.goto(`/upstream-services/${serviceName}`);

    // 2. Click Settings tab (outer)
    await page.getByRole('tab', { name: 'Settings' }).click();

    // 3. Click Policies tab (inner ServiceEditor tab)
    // There might be a slight delay for ServiceEditor to load
    await page.getByRole('tab', { name: 'Policies' }).click();

    // 4. Verify "Quick Visibility Control" is present
    await expect(page.getByText('Quick Visibility Control')).toBeVisible();

    // 5. Verify Tools are listed
    // Using getByText might be ambiguous if multiple elements have the text.
    // Use locator with text.
    await expect(page.locator('label', { hasText: 'tool_a' })).toBeVisible();
    await expect(page.locator('label', { hasText: 'tool_b' })).toBeVisible();
    await expect(page.locator('label', { hasText: 'tool_c' })).toBeVisible();

    // 6. Toggle "tool_b" OFF (Hide it)
    // Default is Allow All. So unchecking it should add a Deny rule.
    // We target the checkbox associated with the label.
    // Shadcn checkbox structure: Button[role=checkbox][id=item-tool_b]
    // Label[for=item-tool_b]
    const checkboxB = page.locator('#item-tool_b');
    await expect(checkboxB).toBeChecked(); // Should be checked by default
    await checkboxB.uncheck();

    // 7. Verify visual feedback (line-through on label)
    // The label has a class that adds line-through.
    const labelB = page.locator('label[for="item-tool_b"]');
    await expect(labelB).toHaveClass(/line-through/);

    // 8. Save
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // 9. Verify Toast or success message
    // Use .first() to handle potential duplicate text in toast title/description or multiple toasts
    await expect(page.getByText('Service Updated').first()).toBeVisible();

    // 10. Reload and Verify Persistence
    await page.reload();

    // Navigate back to Policies
    await page.getByRole('tab', { name: 'Settings' }).click();
    await page.getByRole('tab', { name: 'Policies' }).click();

    // Verify tool_b is unchecked
    await expect(page.locator('#item-tool_b')).not.toBeChecked();
    // Verify tool_a is checked
    await expect(page.locator('#item-tool_a')).toBeChecked();

    // 11. Verify Backend State directly
    const response = await request.get(`/api/v1/services/${serviceName}`);
    const data = await response.json();
    const service = data.service || data;
    const policy = service.tool_export_policy || service.toolExportPolicy;

    expect(policy).toBeDefined();
    // We expect one rule for tool_b
    expect(policy.rules).toHaveLength(1);
    // Unchecked "tool_b" should create a rule: nameRegex: "^tool_b$", action: UNEXPORT (2)
    // Note: Our code escapes regex. ^tool_b$ is expected.
    // Regex matching might need to handle escaping if 'tool_b' had special chars, but here it's simple.
    expect(policy.rules[0].name_regex).toBe('^tool_b$');
    expect(policy.rules[0].action).toBe("UNEXPORT");
  });
});
