import { test, expect } from '@playwright/test';

test.describe('Service Configuration Editor', () => {
  test('should allow editing environment variables for command line service', async ({ page }) => {
    // Navigate to services page
    await page.goto('/services');
    await expect(page).toHaveTitle(/MCPAny Manager/);

    // Open "New Service" sheet
    await page.getByRole('button', { name: 'Add Service' }).click();
    await expect(page.getByText('New Service')).toBeVisible();

    // Select "Command Line" type
    // Depending on the implementation of Select in Shadcn UI, it might need specific steps.
    // The SelectTrigger has text "Select type" initially.
    await page.getByRole('combobox').click();
    await page.getByRole('option', { name: 'Command Line' }).click();

    // Fill command
    await page.getByLabel('Command').fill('echo hello');

    // EnvVarEditor should be visible now
    await expect(page.locator('label', { hasText: 'Environment Variables' })).toBeVisible();

    // Add a variable
    await page.getByRole('button', { name: 'Add Variable' }).click();

    // Fill Key and Value
    await page.getByPlaceholder('KEY').fill('TEST_ENV');
    await page.getByPlaceholder('VALUE').fill('test_value');

    // Take screenshot of the editor
    await page.screenshot({ path: '.audit/ui/2025-02-20/env_var_editor.png' });

    // Verify inputs
    await expect(page.getByPlaceholder('KEY')).toHaveValue('TEST_ENV');
    await expect(page.getByPlaceholder('VALUE')).toHaveValue('test_value');

    // Close
    await page.getByRole('button', { name: 'Cancel' }).click();
  });
});
