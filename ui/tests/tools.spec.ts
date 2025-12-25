
import { test, expect } from '@playwright/test';

test('Tools page lists tools, filters, and allows toggling', async ({ page }) => {
  await page.goto('/tools');

  await expect(page.getByRole('heading', { name: 'Tools' })).toBeVisible();

  // Check list
  await expect(page.getByText('stripe_charge')).toBeVisible();
  await expect(page.getByText('get_user')).toBeVisible();

  // Test Filter
  await page.getByPlaceholder('Search tools...').fill('stripe');
  await expect(page.getByText('stripe_charge')).toBeVisible();
  await expect(page.getByText('get_user')).not.toBeVisible();

  // Clear filter
  await page.getByPlaceholder('Search tools...').fill('');

  // Test Toggle
  // Assuming 'search_docs' is initially disabled (based on mock)
  const toggle = page.locator('tr:has-text("search_docs") button[role="switch"]');
  await expect(toggle).not.toBeChecked();

  await toggle.click();
  await expect(toggle).toBeChecked();
  await expect(page.locator('tr:has-text("search_docs")')).toContainText('Active');
});
