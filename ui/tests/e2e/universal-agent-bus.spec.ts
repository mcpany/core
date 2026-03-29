import { test, expect } from '@playwright/test';
import { seedUser } from './test-data';

test.describe('Universal Agent Bus Page', () => {
  let requestContext;

  test.beforeEach(async ({ page, request }) => {
    requestContext = request;
    await seedUser(requestContext, 'admin');
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'admin');
    await page.click('button[type="submit"]');
    await page.waitForURL('/');
    await page.goto('/universal-agent-bus');
  });

  test('should display the Universal Agent Bus dashboard', async ({ page }) => {
    await expect(page.locator('h1').filter({ hasText: 'Universal Agent Bus' })).toBeVisible();
    await expect(page.getByText('Recursive Context Dashboard')).toBeVisible();
    await expect(page.getByText('Multi-Agent Session Timeline')).toBeVisible();
    await expect(page.getByText('Unified Discovery Manager')).toBeVisible();
    await expect(page.getByText('Lazy-MCP Tool Search Dashboard')).toBeVisible();
    await expect(page.getByText('Agent Chain Tracer (A2A)')).toBeVisible();
  });
});
