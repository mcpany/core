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
      'Lazy-MCP Tool Search Dashboard'
    ];

    for (const card of cards) {
      await expect(page.locator('.text-sm.font-medium', { hasText: card })).toBeVisible();
    }

    await expect(page.locator('.text-xl.font-semibold', { hasText: 'Agent Chain Tracer (A2A)' })).toBeVisible();
  });

  test('should display seeded traces on the Agent Chain Tracer timeline', async ({ page, request }) => {
    // 1. Seed the database with trace data
    const response = await request.post('/api/v1/debug/traces', {
      data: {}
    });
    expect(response.ok()).toBeTruthy();

    // 2. Navigate to the page
    await page.goto('/universal-agent-bus');

    // 3. Verify the timeline elements appear instead of the empty state
    await expect(page.locator('text=orchestrator-task')).toBeVisible();
    await expect(page.locator('text=search-tool')).toBeVisible();

    // 4. Expand one of the trace elements to verify details
    const firstTraceRow = page.locator('.relative.group').first();
    await firstTraceRow.click();

    // 5. Verify the formatted JSON payload exists in the expanded view
    await expect(page.locator('pre code')).toContainText('query');
  });
});
