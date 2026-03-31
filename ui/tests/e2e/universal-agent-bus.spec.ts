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
      'Local Pairing Portal',
      'CRDT Shard Health Explorer',
      'Agent Chain Tracer (A2A)'
    ];

    for (const card of cards) {
      // Use a more general locator that handles both CardTitle (.text-sm.font-medium)
      // and the custom AgentChainTracer title (.text-xl.font-semibold)
      await expect(page.getByText(card)).toBeVisible();
    }
  });
});
