
import { test, expect } from '@playwright/test';

test('Global Search (Cmd+K) should open and display dynamic content', async ({ page }) => {
  // Navigate to the dashboard
  await page.goto('/');

  // Wait for the page to load using a more specific selector
  await expect(page.locator('h2:has-text("Dashboard")')).toBeVisible();

  // Simulate Cmd+K to open the search
  await page.keyboard.press('Meta+k');

  // Verify the dialog is open
  await expect(page.locator('div[role="dialog"]')).toBeVisible();
  await expect(page.locator('input[placeholder="Type a command or search..."]')).toBeVisible();

  // Wait for data to load

  // Check for Suggestions
  await expect(page.getByText('Suggestions')).toBeVisible();

  // Check for dynamic content
  // Services
  await expect(page.getByText('weather-service').first()).toBeVisible();

  // Tools
  await expect(page.getByText('get_weather').first()).toBeVisible();

  // Resources
  await expect(page.getByText('notes.txt').first()).toBeVisible();

  // Prompts
  await expect(page.getByText('summarize_file').first()).toBeVisible();

  // Type in the search box to filter
  await page.fill('input[placeholder="Type a command or search..."]', 'weather');

  // Verify filtering works
  await expect(page.getByText('weather-service').first()).toBeVisible();
  await expect(page.getByText('get_weather').first()).toBeVisible();
  await expect(page.getByText('local-files')).not.toBeVisible();

  // Take a screenshot for audit
  await page.screenshot({ path: '.audit/ui/2025-05-18/global_search_audit.png' });
});
