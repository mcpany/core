import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser, seedCollection, cleanupCollection } from './test-data';

test.describe('Universal Agent Bus', () => {
  test.beforeEach(async ({ request, page }) => {
      await seedCollection('mcpany-system', request);
      await seedUser(request, "e2e-admin-uab");
      // Login
      await page.goto('/login');
      await page.waitForLoadState('networkidle');
      await page.fill('input[name="username"]', "e2e-admin-uab");
      await page.fill('input[name="password"]', 'password');
      await Promise.all([
        page.waitForURL('/', { timeout: 30000 }),
        page.click('button[type="submit"]', { force: true })
      ]);
  });

  test.afterEach(async ({ request }) => {
      await cleanupCollection('mcpany-system', request);
  });

  test('should load the dashboard and display all feature cards', async ({ page }) => {
    // 1. Navigate directly to the Universal Agent Bus page after login
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
    // Note: use wait condition or timeout because ws takes a moment to connect
    await expect(page.getByText('orchestrator-task').first()).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('search-tool').first()).toBeVisible({ timeout: 10000 });

    // 4. Expand one of the trace elements to verify details
    const firstTraceRow = page.locator('.relative.group').first();
    await firstTraceRow.click();

    // 5. Verify the formatted JSON payload exists in the expanded view
    await expect(page.locator('pre code').first()).toContainText('query', { timeout: 10000 });
  });
});
