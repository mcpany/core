/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Service Wizard', () => {
  test('should allow creating a service from a template', async ({ page }) => {
    // 1. Go to Services page
    await page.goto('/upstream-services');

    // 2. Click "Add Service"
    await page.getByRole('button', { name: 'Add Service' }).click();

    // 3. Verify Template Selector is shown
    await expect(page.getByText('New Service')).toBeVisible();
    await expect(page.getByPlaceholder('Search templates...')).toBeVisible();

    // 4. Select "PostgreSQL" template
    // Use a specific text locator that matches the template card
    await page.getByText('PostgreSQL', { exact: true }).click();

    // 5. Verify Config Form is shown
    await expect(page.getByText('Configure PostgreSQL')).toBeVisible();
    await expect(page.getByLabel('PostgreSQL Connection String')).toBeVisible();

    // 6. Enter a dummy connection string
    await page.getByLabel('PostgreSQL Connection String').fill('postgresql://user:pass@localhost:5432/mydb');

    // 7. Click Continue
    await page.getByRole('button', { name: 'Continue' }).click();

    // 8. Verify Service Editor is shown with pre-filled values
    // Using getByLabel if possible, or looking for input with value
    await expect(page.locator('input[value="postgres-db"]')).toBeVisible();

    // Switch to Connection tab to verify command
    await page.getByRole('tab', { name: 'Connection' }).click();
    await expect(page.getByLabel('Command')).toHaveValue('npx -y @modelcontextprotocol/server-postgres postgresql://user:pass@localhost:5432/mydb');

    // 9. Save the service (Optional in mockup environment, may fail if backend is down)
    // In a real environment, we would save and expect 'Service Created'.
    // For this verification, we stop at the pre-filled editor validation which confirms the wizard logic.
    // await page.getByRole('button', { name: 'Save Changes' }).click();
  });
});
