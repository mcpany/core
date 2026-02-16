import { test, expect } from '@playwright/test';

test.describe('Bulk Service Import', () => {
  const uniqueId = Math.random().toString(36).substring(7);
  const service1 = `test-service-1-${uniqueId}`;
  const service2 = `test-service-2-${uniqueId}`;

  test.beforeEach(async ({ page }) => {
    await page.goto('/upstream-services');
  });

  test('should complete bulk import via wizard', async ({ page }) => {
    // Open Dialog
    const importBtn = page.getByRole('button', { name: 'Bulk Import' });
    await expect(importBtn).toBeVisible();
    await importBtn.click();

    const dialog = page.getByRole('dialog', { name: 'Bulk Service Import' });
    await expect(dialog).toBeVisible();

    // Step 1: Input
    await expect(dialog.getByText('Paste JSON / File')).toBeVisible();

    // Paste JSON
    const json = JSON.stringify([
      { name: service1, httpService: { address: "http://example.com" } },
      { name: service2, httpService: { address: "http://example.org" } }
    ]);
    await dialog.locator('textarea').fill(json);

    // Click Next
    const nextBtn = dialog.getByRole('button', { name: 'Next' });
    await expect(nextBtn).toBeEnabled();
    await nextBtn.click();

    // Step 2: Validate
    await expect(dialog.getByText('Review & Validate')).toBeVisible();
    await expect(dialog.getByText(service1)).toBeVisible();
    await expect(dialog.getByText(service2)).toBeVisible();

    // Wait for Re-validate button to be enabled (meaning checking done)
    const revalidateBtn = dialog.getByRole('button', { name: 'Re-validate' });
    await expect(revalidateBtn).toBeEnabled();

    // Check checkboxes. Radix checkbox is tricky. Use locator for role=checkbox inside table.
    // The first one is "Select All" in the header.
    const checkboxes = dialog.locator('button[role="checkbox"]');
    const selectAll = checkboxes.first();

    // Force click to ensure checked state (if not already checked)
    // Radix checkbox has data-state="checked" or "unchecked"
    if (await selectAll.getAttribute('data-state') !== 'checked') {
        await selectAll.click();
    }

    // Click Import
    const importSelectedBtn = dialog.getByRole('button', { name: 'Import Selected' });
    await expect(importSelectedBtn).toBeEnabled();
    await importSelectedBtn.click();

    // Step 3: Wait for completion
    // Check for summary
    await expect(dialog.getByText('Import Complete')).toBeVisible({ timeout: 10000 });

    // Close
    await dialog.getByRole('button', { name: 'Close' }).first().click();

    // Verify dialog closed
    await expect(dialog).not.toBeVisible();

    // Verify services in list
    // Might need reload? The onImportSuccess calls fetchServices, so list should update.
    await expect(page.getByText(service1)).toBeVisible();
    await expect(page.getByText(service2)).toBeVisible();
  });
});
