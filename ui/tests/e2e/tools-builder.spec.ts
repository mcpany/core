import { test, expect } from '@playwright/test';

test.describe('Tools Builder', () => {
  test('should allow building arguments visually and sync to JSON', async ({ page }) => {
    // Navigate to Tools page
    await page.goto('/tools');

    // Wait for the tool list to load
    // Filter for the specific wttr.in tool which has 'location' argument
    const row = page.locator('tr').filter({ hasText: 'wttrin' }).filter({ hasText: 'get_weather' }).first();
    await expect(row).toBeVisible({ timeout: 30000 });

    // Click Inspect
    await row.getByRole('button', { name: 'Inspect' }).click();

    // Verify Dialog Open
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    // Verify Builder Tab is active
    const builderTab = dialog.getByRole('tab', { name: 'Builder' });
    await expect(builderTab).toBeVisible();
    await expect(builderTab).toHaveAttribute('data-state', 'active');

    // Find input for location (wttr.in get_weather has location)
    // Label should be "location"
    const locationInput = dialog.getByLabel('location');
    await expect(locationInput).toBeVisible();

    // Type "Paris"
    await locationInput.fill('Paris');

    // Switch to JSON tab
    // Locate the arguments tab list
    const argTabs = builderTab.locator('..');
    await argTabs.getByRole('tab', { name: 'JSON' }).click();

    // Verify JSON content
    const jsonInput = dialog.locator('textarea#args');
    await expect(jsonInput).toBeVisible();
    // Verify it contains location: Paris
    await expect(jsonInput).toHaveValue(/"location": "Paris"/);

    // Switch back to Builder
    await argTabs.getByRole('tab', { name: 'Builder' }).click();

    // Verify input still has "Paris" (remounted correctly)
    await expect(locationInput).toHaveValue('Paris');

    // Close dialog
    // Use first() to pick the footer button (which appeared first in list), avoiding the X button
    await dialog.getByRole('button', { name: 'Close' }).first().click();

    // Reopen inspect to verify state reset
    await row.getByRole('button', { name: 'Inspect' }).click();

    // Verify input is empty (or default)
    // wttr.in location has no default, so it should be empty.
    await expect(locationInput).toHaveValue('');

    // Verify JSON is reset (no "Paris"), but might have defaults like "lang": "en"
    await argTabs.getByRole('tab', { name: 'JSON' }).click();
    await expect(jsonInput).not.toHaveValue(/Paris/);
    await expect(jsonInput).toHaveValue(/"lang": "en"/);
  });
});
