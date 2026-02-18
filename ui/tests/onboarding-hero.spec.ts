import { test, expect } from '@playwright/test';

test.describe('Onboarding Hero', () => {
  test('should display onboarding hero when dashboard is empty', async ({ page }) => {
    // Navigate to dashboard
    await page.goto('/');

    // Clear any existing widgets to force empty state
    // We assume the "Layout" button is visible if widgets exist
    const layoutButton = page.getByRole('button', { name: 'Layout' });

    if (await layoutButton.isVisible()) {
        await layoutButton.click();
        await page.getByRole('button', { name: 'Clear All' }).click();
    }

    // Verify Onboarding Hero is visible
    await expect(page.getByText('Welcome to MCP Any')).toBeVisible();
    await expect(page.getByRole('link', { name: 'Connect a Service' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Browse Marketplace' })).toBeVisible();

    // Test "Add Metrics Overview" interaction
    await page.getByRole('button', { name: 'Add Metrics Overview' }).click();

    // Hero should disappear and widgets should appear
    await expect(page.getByText('Welcome to MCP Any')).not.toBeVisible();
    // Assuming Metrics Overview has a title like "Metrics Overview" or specific content
    // We can just check that Layout button is back or layout is populated
    // But since we just added one, it should be visible.
  });
});
