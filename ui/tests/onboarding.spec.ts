import { test, expect } from '@playwright/test';
import { seedCollection } from './e2e/test-data';

test.describe('Onboarding Flow', () => {
  // Test 1: Empty State
  test('shows onboarding hero when no services exist', async ({ page }) => {
    // Note: We rely on the environment starting clean or this test running in isolation.
    // If DB has data, this test will fail.
    // Ideally, we would call an API to clear data here.

    await page.goto('/');

    // Check for "Welcome to MCP Any"
    await expect(page.getByText('Welcome to MCP Any')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Connect Service')).toBeVisible();
    await expect(page.getByText('Browse Marketplace')).toBeVisible();

    // Ensure Dashboard Grid is NOT visible
    // "Metrics Overview" is part of the default dashboard layout
    await expect(page.getByText('Metrics Overview')).not.toBeVisible();
  });

  // Test 2: Populated State
  test('shows dashboard when services exist', async ({ page, request }) => {
    // Seed a service
    await seedCollection('mcpany-system', request);

    await page.goto('/');

    // Check for Dashboard Grid
    await expect(page.getByText('Metrics Overview')).toBeVisible({ timeout: 10000 });

    // Ensure Onboarding Hero is NOT visible
    await expect(page.getByText('Welcome to MCP Any')).not.toBeVisible();
  });
});
