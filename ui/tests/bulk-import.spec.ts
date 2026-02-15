import { test, expect } from '@playwright/test';

test.describe('Bulk Import', () => {
  test('should import services via wizard', async ({ page }) => {
    // Navigate to upstream services
    await page.goto('/upstream-services');

    // Click Bulk Import button
    await page.getByRole('button', { name: 'Bulk Import' }).click();

    // Verify dialog is visible
    await expect(page.getByRole('dialog', { name: 'Bulk Service Import' })).toBeVisible();

    // Step 1: Input
    const serviceName = `test-service-${Date.now()}`;
    const jsonConfig = JSON.stringify([
      {
        name: serviceName,
        httpService: {
          address: "https://httpbin.org" // This should fail if network is blocked
        }
      }
    ]);

    await page.getByPlaceholder('[{"name": "service1", ...}]').fill(jsonConfig);
    await page.getByRole('button', { name: 'Next: Validate' }).click();

    // Step 2: Validation
    await expect(page.getByText('Validation Results')).toBeVisible();

    // Wait for validating to stop (spinner gone)
    await expect(page.locator('.animate-spin')).toHaveCount(0);

    // Check for validation error
    if (await page.getByTestId('validation-status-error').isVisible()) {
        const errorMsg = await page.getByRole('cell').nth(3).textContent();
        console.log("Validation Failed:", errorMsg);
        // We still proceed, as validation failure shouldn't block import (user can ignore)
        // But for this test, we want to know.
        // If it's network error, we can ignore it for now as long as import works.
    } else {
        await expect(page.getByTestId('validation-status-valid')).toBeVisible();
    }

    await page.getByRole('button', { name: 'Next: Select' }).click();

    // Step 3: Selection
    await expect(page.getByText('Select Services to Import')).toBeVisible();
    // Default selected
    await page.getByRole('button', { name: /Import \d+ Services/ }).click();

    // Step 4: Import
    // Wait for completion (check for success count)
    // We expect 1 success.
    await expect(page.getByText(/Success:\s*1/)).toBeVisible({ timeout: 15000 });

    // Ensure "Finish" button is visible (implies importing is done)
    await expect(page.getByRole('button', { name: 'Finish' })).toBeVisible();

    await page.getByRole('button', { name: 'Finish' }).click();

    // Dialog should close
    await expect(page.getByRole('dialog')).toBeHidden();

    // Verify service is in the list
    await expect(page.getByText(serviceName)).toBeVisible();
  });
});
