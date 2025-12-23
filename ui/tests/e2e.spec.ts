
import { test, expect } from '@playwright/test';

test.describe('MCP Any UI E2E Tests', () => {
  test('Dashboard loads and displays metrics', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/MCPAny Manager/);
    await expect(page.getByText('Total Requests')).toBeVisible();
    await expect(page.getByText('Active Services')).toBeVisible();
    await expect(page.getByText('Overview')).toBeVisible();
  });

  test('Services page loads and displays service list', async ({ page }) => {
    await page.goto('/services');
    await expect(page.getByRole('heading', { name: 'Services' })).toBeVisible();
    // Check for mock data
    await expect(page.getByText('Payment Gateway')).toBeVisible();
    await expect(page.getByText('User Service')).toBeVisible();
  });

  test('Tools page loads', async ({ page }) => {
    await page.goto('/tools');
    await expect(page.getByRole('heading', { name: 'Tools' })).toBeVisible();
    await expect(page.getByText('get_weather')).toBeVisible();
  });

   test('Middleware page has pipeline', async ({ page }) => {
    await page.goto('/middleware');
    await expect(page.getByRole('heading', { name: 'Middleware Pipeline' })).toBeVisible();
    await expect(page.getByText('Authentication Guard')).toBeVisible();
  });
});
