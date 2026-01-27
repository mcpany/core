import { test, expect } from '@playwright/test';

test('dashboard widget gallery', async ({ page }) => {
  await page.goto('/');

  // Check if "Add Widget" button exists
  await expect(page.getByRole('button', { name: 'Add Widget' })).toBeVisible();

  // Open the sheet
  await page.getByRole('button', { name: 'Add Widget' }).click();

  // Check if gallery is visible
  await expect(page.getByText('Choose a widget to add')).toBeVisible();

  // Add a "Service Health" widget.
  // In the gallery it is called "Service Health".
  // Use filter to be specific.
  const galleryItem = page.locator('.grid > div').filter({ hasText: 'Service Health' }).first();
  await galleryItem.click();

  // Verify we have widgets on the dashboard.
  // The ServiceHealthWidget renders with title "System Health".
  // Since we might have multiple, let's just check that at least one is visible.
  await expect(page.getByText('System Health').first()).toBeVisible();

  // Optionally check count if we started with known state, but default layout might change.
});
