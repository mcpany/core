import { test, expect } from '@playwright/test';

test.describe('Universal Agent Bus Page', () => {
  test('should load the Universal Agent Bus page successfully', async ({ page }) => {
    // We navigate to the Universal Agent Bus page directly
    await page.goto('/universal-agent-bus');

    // Wait for the main heading to be visible
    const heading = page.locator('h2:has-text("Universal Agent Bus")');
    await expect(heading).toBeVisible();

    // Verify the page structure
    await expect(page.locator('text=Active Swarms')).toBeVisible();
    await expect(page.locator('text=Agent Chain Tracer (A2A)')).toBeVisible();
    await expect(page.locator('text=No active agent interactions detected.')).toBeVisible();
  });
});
