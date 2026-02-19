import { test, expect } from '@playwright/test';

test('getting started widget appears on dashboard', async ({ page }) => {
  // Mock the services API to return empty list
  await page.route('/api/v1/services', async route => {
    await route.fulfill({ json: [] });
  });

  // Mock preferences to ensure default layout
  await page.route('/api/v1/user/preferences', async route => {
    await route.fulfill({ status: 404 });
  });

  await page.goto('/');

  // Expect to see the Getting Started widget title
  await expect(page.getByText('Welcome to MCP Any')).toBeVisible();

  // Expect to see the "Quick Start" button
  await expect(page.getByText('Quick Start')).toBeVisible();
});
