import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Universal Agent Bus E2E', () => {
  test.beforeEach(async ({ request, page }) => {
    // Rely on UI-interaction seeding paradigm where possible, but use global state to ensure user exists
    await seedGlobalState(request);

    // Login
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', 'e2e-admin-core');
    await page.fill('input[name="password"]', 'password');
    await Promise.all([
      page.waitForURL('/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
    ]);
    await expect(page).toHaveURL('/', { timeout: 15000 });
  });

  test('should navigate to Universal Agent Bus and view dashboards', async ({ page }) => {
    // Go to the Universal Agent Bus page via URL
    await page.goto('/universal-agent-bus');

    // Wait for main title
    await expect(page.locator('h1')).toContainText('Universal Agent Bus');

    // Verify cards appear
    await expect(page.getByText('Recursive Context Dashboard')).toBeVisible();
    await expect(page.getByText('Multi-Agent Session Timeline')).toBeVisible();
    await expect(page.getByText('Unified Discovery Manager')).toBeVisible();
    await expect(page.getByText('Lazy-MCP Tool Search Dashboard')).toBeVisible();
    await expect(page.getByText('Agent Chain Tracer (A2A)')).toBeVisible();

    // Verify some values
    await expect(page.getByText('Inactive').first()).toBeVisible();
    await expect(page.getByText('0 Sessions')).toBeVisible();
    await expect(page.getByText('0 Transports')).toBeVisible();
    await expect(page.getByText('0 Index Hits')).toBeVisible();
    await expect(page.getByText('0 Handoffs')).toBeVisible();
  });
});
