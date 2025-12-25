
import { test, expect } from '@playwright/test';

test('Dashboard loads and displays metrics', async ({ page }) => {
  await page.goto('/');

  // Expect a title "to contain" a substring.
  await expect(page).toHaveTitle(/MCPAny Manager/);

  // Check for Dashboard header
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();

  // Check for metrics cards
  await expect(page.getByText('Total Requests')).toBeVisible();
  await expect(page.getByText('Active Services')).toBeVisible();
  await expect(page.getByText('Avg Latency')).toBeVisible();
  await expect(page.getByText('Active Users')).toBeVisible();
});
