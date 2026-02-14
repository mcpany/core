import { test, expect } from '@playwright/test';

test.describe('Bulk Service Import Wizard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/upstream-services');
  });

  test('should complete the full import wizard flow', async ({ page }) => {
    // 1. Open Wizard
    await page.getByRole('button', { name: /Bulk Import/i }).click();
    await expect(page.getByRole('dialog')).toContainText('Bulk Service Import');

    // 2. Input JSON
    // Using a random name to avoid conflicts if test runs multiple times
    const serviceName = `google-test-${Date.now()}`;
    const validConfig = [
        {
            name: serviceName,
            httpService: { address: "https://www.google.com" }
        }
    ];
    await page.getByRole('textbox').fill(JSON.stringify(validConfig));
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 3. Validation Step
    await expect(page.getByText('Review & Validate')).toBeVisible();

    // Wait for validation to finish (check for Valid status)
    // The table cell should eventually contain "Valid configuration" or similar
    await expect(page.getByText('Valid configuration')).toBeVisible({ timeout: 15000 });

    // Ensure "Import Selected" is enabled (implies selection)
    const importBtn = page.getByRole('button', { name: /Import Selected/i });
    await expect(importBtn).toBeEnabled();

    // 4. Import
    await importBtn.click();

    // 5. Results
    await expect(page.getByText('Import Complete')).toBeVisible();
    await expect(page.getByText('Successfully processed 1 services')).toBeVisible();

    // 6. Close
    await page.getByRole('button', { name: /Done/i }).click();
    await expect(page.getByRole('dialog')).toBeHidden();

    // 7. Verify in list
    // We filter by name to find the specific row/item
    await expect(page.getByText(serviceName)).toBeVisible();
  });
});
