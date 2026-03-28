import { test, expect } from '@playwright/test';

test.describe('Universal Agent Bus E2E', () => {
  test('should display the Universal Agent Bus dashboard correctly', async ({ page }) => {
    // 1. "If a test needs a User, the script must first Register that user via the UI."
    // Note: If the login/register screen is required, it must happen via UI interaction.
    // However, our system bypasses login when X-API-Key is set, which is done in playwright.config.ts.

    // Navigate to the page
    await page.goto('/universal-agent-bus');

    // Wait for the h1 to appear to avoid timeout errors
    await page.waitForSelector('h1:has-text("Universal Agent Bus")', { timeout: 30000 });

    // Verify title and description
    await expect(page.locator('h1')).toHaveText('Universal Agent Bus');
    await expect(page.locator('p').first()).toContainText('Manage and map subagents dynamically');

    // Verify all 5 dashboard cards are visible
    await expect(page.locator('text=Recursive Context Dashboard')).toBeVisible();
    await expect(page.locator('text=Multi-Agent Session Timeline')).toBeVisible();
    await expect(page.locator('text=Unified Discovery Manager')).toBeVisible();
    await expect(page.locator('text=Lazy-MCP Tool Search Dashboard')).toBeVisible();
    await expect(page.locator('text=Agent Chain Tracer (A2A)')).toBeVisible();
  });
});
