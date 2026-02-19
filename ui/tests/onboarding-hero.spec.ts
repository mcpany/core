import { test, expect } from '@playwright/test';

test.describe('Onboarding Hero', () => {
  test('should display onboarding hero when dashboard is empty', async ({ page }) => {
    // Navigate to dashboard
    await page.goto('/');

    // Wait for the app to load (look for the header or dashboard title)
    await expect(page.getByText('Dashboard')).toBeVisible();

    // Check if we are already in empty state
    const hero = page.getByText('Welcome to MCP Any');
    if (await hero.isVisible()) {
        // Already empty, test passes/proceeds
    } else {
        // Widgets are present, we need to clear them
        const layoutButton = page.getByRole('button', { name: 'Layout' });
        await expect(layoutButton).toBeVisible();
        await layoutButton.click();

        const clearAllButton = page.getByRole('button', { name: 'Clear All' });
        await expect(clearAllButton).toBeVisible();
        await clearAllButton.click();
    }

    // Verify Onboarding Hero is visible
    await expect(page.getByText('Welcome to MCP Any')).toBeVisible();
    await expect(page.getByRole('link', { name: 'Connect a Service' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Browse Marketplace' })).toBeVisible();

    // Test "Add Metrics Overview" interaction
    await page.getByRole('button', { name: 'Add Metrics Overview' }).click();

    // Hero should disappear and widgets should appear
    await expect(page.getByText('Welcome to MCP Any')).not.toBeVisible();
  });
});
