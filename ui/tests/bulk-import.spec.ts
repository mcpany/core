import { test, expect } from '@playwright/test';

test.describe('Bulk Service Import Wizard', () => {
  test('should import services using the wizard', async ({ page }) => {
    // Navigate to services page
    await page.goto('/upstream-services');

    // Click on Bulk Import button
    await page.getByRole('button', { name: 'Bulk Import' }).click();

    // Expect dialog to be visible
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Bulk Service Import' })).toBeVisible();

    // Step 1: Input
    // Verify tabs
    await expect(page.getByRole('tab', { name: 'JSON / YAML' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'URL' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'File Upload' })).toBeVisible();

    // Paste JSON
    const services = [
      {
        name: "test-bulk-1",
        httpService: { address: "https://example.com" }
      },
      {
        name: "test-bulk-2",
        commandLineService: { command: "echo", args: ["hello"] }
      }
    ];
    await page.getByLabel('Configuration Content').fill(JSON.stringify(services));

    // Click Next
    await page.getByRole('button', { name: 'Review & Validate' }).click();

    // Step 2: Validation
    // Wait for validation to complete (table should appear)
    await expect(page.getByRole('heading', { name: 'Validation Results' })).toBeVisible();

    // Check if services are listed
    await expect(page.getByRole('cell', { name: 'test-bulk-1' })).toBeVisible();
    await expect(page.getByRole('cell', { name: 'test-bulk-2' })).toBeVisible();

    // Wait for status. We expect valid because example.com is reachable and echo is available.
    // Note: If backend validation is strict about reachable URLs, this might fail if no internet in backend container.
    // If it fails, it shows Error badge.
    // Let's just check that we can proceed. If invalid, the import button is disabled if 0 selected.
    // But invalid ones are auto-deselected.

    // Let's select them manually just in case? No, disabled if invalid.
    // We assume backend connectivity check passes.
    // If it fails, the test will fail on "Import" click or verification.

    // Click Import (matches "Import 2 Services" or similar)
    await page.getByRole('button', { name: /Import \d+ Services/ }).click();

    // Step 3: Import Progress -> Result
    // Should transition to Result step eventually
    await expect(page.getByRole('heading', { name: 'Import Complete' })).toBeVisible({ timeout: 10000 });

    // Verify success
    await expect(page.getByText(/Successfully imported/)).toBeVisible();

    // Close dialog (target the primary action button, not the X)
    // The X button also has accessible name "Close", so we need to be specific
    await page.getByRole('button', { name: 'Close' }).first().click();

    // Verify dialog closed
    await expect(page.getByRole('dialog')).toBeHidden();

    // Verify services are in the list
    // You might need to reload or wait for fetch
    await expect(page.getByRole('link', { name: 'test-bulk-1' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'test-bulk-2' })).toBeVisible();
  });
});
