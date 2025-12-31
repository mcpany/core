
import { test, expect } from '@playwright/test';

test.describe('MCP Any UI E2E', () => {

  test('Dashboard loads and shows metrics', async ({ page }) => {
    await page.goto('/');
    // Updated title expectation
    await expect(page).toHaveTitle(/MCPAny Manager/);
    await expect(page.locator('h2')).toContainText('Dashboard');
    // Check for metrics cards
    await expect(page.locator('text=Total Requests')).toBeVisible();
    await expect(page.locator('text=System Health')).toBeVisible();
  });

  test('Services page CRUD', async ({ page }) => {
    await page.goto('/services');
    await expect(page.locator('h2')).toContainText('Services');

    // Add Service
    await page.click('button:has-text("Add Service")');
    await expect(page.locator('div[role="dialog"]')).toBeVisible();
    await page.fill('input#name', 'test-service-e2e');
    await page.click('text=Save Changes');

    // Check if added
    await expect(page.locator('text=test-service-e2e')).toBeVisible();
  });

  test('Tools page lists tools and inspects', async ({ page }) => {
    await page.goto('/tools');
    await expect(page.locator('h2')).toContainText('Tools');
    await expect(page.locator('text=get_weather')).toBeVisible();

    // Inspect
    await page.click('button:has-text("Inspect") >> nth=0');
    await expect(page.locator('div[role="dialog"]')).toBeVisible();
    await expect(page.locator('text=Schema')).toBeVisible();
  });

  test('Middleware page drag and drop', async ({ page }) => {
    await page.goto('/middleware');
    await expect(page.locator('h2')).toContainText('Middleware Pipeline');
    await expect(page.locator('text=Active Pipeline')).toBeVisible();
    // Resolving ambiguity by selecting the first occurrence (likely the list item)
    await expect(page.locator('text=Authentication').first()).toBeVisible();
  });

});
