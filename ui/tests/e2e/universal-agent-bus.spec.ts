import { test, expect } from '@playwright/test';

test.describe('Universal Agent Bus', () => {
  test('should load the dashboard and display all feature cards', async ({ page }) => {
    // 1. Perform a real user interaction to navigate to the page
    // Assuming the user is not authenticated or there is a public route
    await page.goto('/');

    // 2. Navigate via sidebar if possible, else go directly to the Universal Agent Bus page
    await page.goto('/universal-agent-bus');

    // Verify title and description are present.
    await expect(page.locator('h1')).toHaveText('Universal Agent Bus');
    await expect(page.getByText('Manage and map subagents dynamically')).toBeVisible();

    // Verify all the feature cards are rendered correctly.
    const cards = [
      'Recursive Context Dashboard',
      'Multi-Agent Session Timeline',
      'Unified Discovery Manager',
      'Lazy-MCP Tool Search Dashboard',
    ];

    for (const card of cards) {
      await expect(page.locator('.text-sm.font-medium', { hasText: card })).toBeVisible();
    }

    // Agent Chain Tracer has a different DOM structure (it uses an h3 CardTitle inside a different layout)
    // so we locate it by role and name instead.
    await expect(page.getByRole('heading', { name: 'Agent Chain Tracer (A2A)' })).toBeVisible();
  });
});
