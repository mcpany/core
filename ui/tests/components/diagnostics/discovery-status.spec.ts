import { test, expect } from '@playwright/test';

test.describe('DiscoveryStatus Component Documentation Types', () => {
    test('renders DiscoveryStatus correctly', async ({ page }) => {
        // Just verify it doesn't crash from the TSDoc change
        await page.goto('/diagnostics');
        await expect(page.getByText('Auto-Discovery Status')).toBeVisible();
    });
});
