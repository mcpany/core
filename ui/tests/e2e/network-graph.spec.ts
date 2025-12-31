import { test, expect } from '@playwright/test';

test('network topology loads and displays nodes', async ({ page }) => {
  await page.goto('/network');

  // Wait for the graph to load
  await expect(page.getByText('Network Graph')).toBeVisible();
  await expect(page.getByText('Live topology of MCP services and tools.')).toBeVisible();

  // Check for presence of key nodes (from mock data)
  // React Flow renders nodes as divs with text
  await expect(page.getByText('MCP Any Core')).toBeVisible();
  await expect(page.getByText('weather-service')).toBeVisible();
  await expect(page.getByText('Web Dashboard (Admin)')).toBeVisible();

  // Test interaction
  await page.getByText('weather-service').click();

  // Sheet should open
  await expect(page.getByText('ID: srv-1')).toBeVisible();
  await expect(page.getByText('Operational Status')).toBeVisible();
});
