import { test, expect } from '@playwright/test';

test.describe('Diagnostics Page', () => {
  test('should display discovery status', async ({ page }) => {
    // Navigate to diagnostics page
    await page.goto('/diagnostics');

    // Check for Discovery Status card title
    await expect(page.getByText('Discovery Status')).toBeVisible();

    // Check if refresh button exists (using the icon class or button role)
    // We look for a button inside the card header
    const cardHeader = page.locator('.card-header').filter({ hasText: 'Discovery Status' });
    // Note: class names like .card-header depend on implementation details or Tailwind classes.
    // Better to use accessible locators.
    // The component uses <CardHeader> which renders a div with classes.
    // We can look for the button near the title.

    // Check for "Discovery Status" text
    const title = page.getByText('Discovery Status', { exact: false });
    await expect(title).toBeVisible();

    // Check for the "System Diagnostics" page title
    await expect(page.getByRole('heading', { name: 'System Diagnostics' })).toBeVisible();

    // Verify System Health component is also present
    await expect(page.getByText('Healthy').or(page.getByText('Degraded')).or(page.getByText('Unknown'))).toBeVisible();
  });
});
